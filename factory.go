package properties

import (
	"bytes"
	"context"
	"fmt"
	"github.com/araddon/dateparse"
	"gopkg.in/yaml.v2"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	// ThePropertyFactory is primary property factory for common use cases
	ThePropertyFactory = &DefaultPropertyFactory{}

	// ThePropertiesFactory is primary properties collection factory for common use cases
	ThePropertiesFactory = &DefaultPropertiesFactory{PropFactory: ThePropertyFactory}
)

// CustomCreatorFunc is provided in factory for custom property creation use cases
type CustomCreatorFunc func(context.Context, string, interface{}, ...interface{}) (Property, bool, error)

// CustomCreatorHandler is provided in factory for custom property creation use cases, overrides CustomCreatorFunc
type CustomCreatorHandler interface {
	FromText(context.Context, string, string, ...interface{}) (Property, bool, error)
	FromAny(context.Context, string, interface{}, ...interface{}) (Property, bool, error)
}

// AfterCreateHookFunc is provided in factory to allow wrapping properties
type AfterCreateHookFunc func(context.Context, Property, ...interface{}) (Property, bool, error)

// AfterCreateHook is provided in factory to allow wrapping properties, overrides AfterCreateHookFunc
type AfterCreateHook interface {
	AfterCreate(context.Context, Property, ...interface{}) (Property, bool, error)
}

// PropertyFactory creates property instances
type PropertyFactory interface {
	FromText(ctx context.Context, name string, value string, options ...interface{}) (Property, bool, error)
	FromAny(ctx context.Context, name string, value interface{}, options ...interface{}) (Property, bool, error)
}

// Factory creates Properties instances
type Factory interface {
	PropertyFactory(context.Context) PropertyFactory
	EmptyMutable(context.Context, ...interface{}) MutableProperties
	ImmutableFromStringMap(context.Context, map[string]interface{}, AllowAddFunc, ...interface{}) (Properties, uint, error)
	MutableFromStringMap(context.Context, map[string]interface{}, AllowAddFunc, ...interface{}) (MutableProperties, uint, error)
	MutableFromFrontMatter(context.Context, []byte, bool, AllowAddFunc, AllowAddTextFunc, ...interface{}) ([]byte, MutableProperties, uint, error)
}

// DefaultPropertyFactory is the default instance
type DefaultPropertyFactory struct {
	CustomCreatorFunc   CustomCreatorFunc
	CustomCreator       CustomCreatorHandler
	AfterCreateHookFunc AfterCreateHookFunc
	AfterCreate         AfterCreateHook
}

// FromAny takes a property name and a value, then creates a typed Property from it
// A CustomCreatorFunc or CustomCreator may be passed in options to handle unknown (custom) property types
func (f *DefaultPropertyFactory) FromAny(ctx context.Context, name string, v interface{}, options ...interface{}) (Property, bool, error) {
	switch value := v.(type) {
	case string:
		return f.afterSuccessfulCreate(ctx, &DefaultTextProperty{PropertyName(name), value}, options...)
	case []string:
		return f.afterSuccessfulCreate(ctx, &DefaultTextListProperty{PropertyName(name), value}, options...)
	case time.Time:
		return f.afterSuccessfulCreate(ctx, &DefaultDateTimeProperty{PropertyName(name), value}, options...)
	case bool:
		return f.afterSuccessfulCreate(ctx, &DefaultFlagProperty{PropertyName(name), value}, options...)
	case int:
		return f.afterSuccessfulCreate(ctx, &DefaultCardinalProperty{PropertyName(name), int64(value)}, options...)
	case int64:
		return f.afterSuccessfulCreate(ctx, &DefaultCardinalProperty{PropertyName(name), value}, options...)
	default:
		return f.handleUnknownType(ctx, name, v, options...)
	}
}

// FromText takes a property name and attempts to create typed properties from a text value
func (f *DefaultPropertyFactory) FromText(ctx context.Context, name string, value string, options ...interface{}) (Property, bool, error) {
	if flag, err := strconv.ParseBool(value); err == nil {
		return f.FromAny(ctx, name, flag, options...)
	}

	if dateTime, err := dateparse.ParseAny(value); err == nil {
		return f.FromAny(ctx, name, dateTime, options...)
	}

	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		return f.FromAny(ctx, name, number, options...)
	}

	return f.FromAny(ctx, name, value, options...)
}

func (f *DefaultPropertyFactory) afterSuccessfulCreate(ctx context.Context, property Property, options ...interface{}) (Property, bool, error) {
	if f.AfterCreate != nil {
		return f.AfterCreate.AfterCreate(ctx, property, options...)
	}
	if f.AfterCreateHookFunc != nil {
		return f.AfterCreateHookFunc(ctx, property, options...)
	}

	return property, true, nil
}

func (f *DefaultPropertyFactory) handleUnknownType(ctx context.Context, name string, value interface{}, options ...interface{}) (Property, bool, error) {
	for _, option := range options {
		if fn, ok := option.(CustomCreatorFunc); ok {
			return fn(ctx, name, value, options...)
		}
		if instance, ok := option.(CustomCreatorHandler); ok {
			return instance.FromAny(ctx, name, value, options...)
		}
	}

	if f.CustomCreator != nil {
		return f.CustomCreator.FromAny(ctx, name, value)
	}
	if f.CustomCreatorFunc != nil {
		return f.CustomCreatorFunc(ctx, name, value)
	}
	return nil, false, fmt.Errorf("Unable to add %q property, type %T is not known: %+v", name, value, value)
}

// DefaultPropertiesFactory is the default properties factory
type DefaultPropertiesFactory struct {
	PropFactory PropertyFactory
}

// PropertyFactory returns the factory that is used to produce property instances
func (f *DefaultPropertiesFactory) PropertyFactory(context.Context) PropertyFactory {
	return f.PropFactory
}

// EmptyMutable returns an empty but mutable properties instance
func (f *DefaultPropertiesFactory) EmptyMutable(ctx context.Context, options ...interface{}) MutableProperties {
	return newDefaultProperties(ctx, f.PropertyFactory(ctx), options...)
}

// ImmutableFromStringMap returns a new Properties instance filled with the given items
func (f *DefaultPropertiesFactory) ImmutableFromStringMap(ctx context.Context, items map[string]interface{}, allow AllowAddFunc, options ...interface{}) (Properties, uint, error) {
	return f.fromStringMap(ctx, items, allow, options...)
}

// MutableFromStringMap returns a new Properties instance filled with the given items
func (f *DefaultPropertiesFactory) MutableFromStringMap(ctx context.Context, items map[string]interface{}, allow AllowAddFunc, options ...interface{}) (MutableProperties, uint, error) {
	return f.fromStringMap(ctx, items, allow, options...)
}

// MutableFromFrontMatter returns a new Properties instance from content that looks like a markdown file with front matter
func (f *DefaultPropertiesFactory) MutableFromFrontMatter(ctx context.Context, content []byte, smartParseFM bool, allow AllowAddFunc, allowText AllowAddTextFunc, options ...interface{}) (bodyWithoutFrontMatter []byte, frontMatter MutableProperties, count uint, err error) {
	return f.fromYAMLFrontMatter(ctx, content, smartParseFM, allow, allowText, options...)
}

// FromStringMap returns a new properties instance based on a text map
func (f *DefaultPropertiesFactory) fromStringMap(ctx context.Context, items map[string]interface{}, allow AllowAddFunc, options ...interface{}) (MutableProperties, uint, error) {
	if items == nil {
		return nil, 0, fmt.Errorf("items is Nil")
	}

	props := f.EmptyMutable(ctx, options...)
	count, err := props.AddMap(ctx, items, allow, options...)
	return props, count, err
}

// fromYAMLFrontMatter will convert an input byte array like ---<stuff>---\n<body> into v as YAML and <body> as return value
func (f *DefaultPropertiesFactory) fromYAMLFrontMatter(ctx context.Context, b []byte, smartParseFM bool, allow AllowAddFunc, allowText AllowAddTextFunc, options ...interface{}) ([]byte, MutableProperties, uint, error) {
	buf := bytes.NewBuffer(b)

	var insideFrontMatter bool
	var yamlStartIndex int
	var yamlEndIndex int

	for {
		line, err := buf.ReadString('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, 0, err
		}

		if strings.TrimSpace(line) != "---" {
			continue
		}

		if !insideFrontMatter {
			insideFrontMatter = true
			yamlStartIndex = len(b) - buf.Len()
		} else {
			yamlEndIndex = len(b) - buf.Len()
			break
		}
	}

	// if we get to here and we're not inside front matter then the entire string is body
	if !insideFrontMatter {
		return b, nil, 0, nil
	}

	if insideFrontMatter && yamlEndIndex == 0 {
		return nil, nil, 0, fmt.Errorf("Unexplained front matter parser error; insideFrontMatter: %v, yamlStartIndex: %v, yamlEndIndex: %v", insideFrontMatter, yamlStartIndex, yamlEndIndex)
	}

	var props MutableProperties
	var count uint
	var err error

	if smartParseFM {
		items := make(map[string]string)
		err := yaml.Unmarshal(b[yamlStartIndex:yamlEndIndex], items)
		if err != nil {
			return nil, nil, 0, nil
		}
		props = f.EmptyMutable(ctx, options...)
		count, err = props.AddTextMap(ctx, items, allowText, options...)
	} else {
		items := make(map[string]interface{})
		err := yaml.Unmarshal(b[yamlStartIndex:yamlEndIndex], items)
		if err != nil {
			return nil, nil, 0, nil
		}
		props, count, err = f.fromStringMap(ctx, items, allow, options...)
	}

	return bytes.TrimSpace(b[yamlEndIndex:]), props, count, err
}
