package properties

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Properties manages a group of strongly typed properties, immutable
type Properties interface {
	List(context.Context) []Property
	Map(context.Context) map[string]interface{}
	Named(context.Context, PropertyName) (Property, bool)
	Filter(context.Context, func(context.Context, Property) bool) []Property
	Range(context.Context, func(context.Context, Property) bool)
	RangeNameValue(context.Context, func(context.Context, string, interface{}))
	Size(context.Context) uint
	Write(context.Context, io.Writer, func(context.Context, io.Writer, Property) (bool, error)) error
}

// MutableProperties adds mutability to Properties
type MutableProperties interface {
	Properties
	AddMap(context.Context, map[string]interface{}, ...interface{}) (uint, error)
	AddTextMap(context.Context, map[string]string, ...interface{}) (uint, error)
	AddAny(context.Context, string, interface{}, ...interface{}) (Property, bool, error)
	AddText(context.Context, string, string, ...interface{}) (Property, bool, error)
	AddProperty(context.Context, Property, ...interface{}) error
	Delete(context.Context, PropertyName, ...interface{}) (bool, error)
	DeleteProperty(context.Context, Property, ...interface{}) (bool, error)
}

// Default is the default properties implementation (supports mutability)
type Default struct {
	pf          PropertyFactory
	syncMap     sync.Map
	syncMapSize uint
}

func newDefaultProperties(ctx context.Context, pf PropertyFactory, options ...interface{}) *Default {
	return &Default{pf: pf}
}

// AddMap adds all the items in the given map
func (p *Default) AddMap(ctx context.Context, items map[string]interface{}, options ...interface{}) (uint, error) {
	if items == nil {
		return 0, fmt.Errorf("items is Nil in properties.Default.AddMap")
	}

	var count uint
	for name, value := range items {
		_, ok, err := p.AddAny(ctx, name, value, options...)
		if err != nil {
			return count, err
		}
		if ok {
			count++
		}
	}

	return count, nil
}

// AddTextMap adds all the items in the given map by trying to "smart parse" the text
func (p *Default) AddTextMap(ctx context.Context, items map[string]string, options ...interface{}) (uint, error) {
	if items == nil {
		return 0, fmt.Errorf("items is Nil in properties.Default.AddTextMap")
	}

	var count uint
	for name, value := range items {
		_, ok, err := p.AddText(ctx, name, value, options...)
		if err != nil {
			return count, err
		}
		if ok {
			count++
		}
	}

	return count, nil
}

// AddText adds a single named property of a text value by "smart parsing" the value type
func (p *Default) AddText(ctx context.Context, name string, value string, options ...interface{}) (Property, bool, error) {
	prop, ok, err := p.pf.FromText(ctx, name, value, options...)
	if err != nil {
		return nil, false, err
	}

	if ok {
		p.AddProperty(ctx, prop)
	}
	return prop, ok, nil
}

// AddAny adds a single named property of any value type
func (p *Default) AddAny(ctx context.Context, name string, value interface{}, options ...interface{}) (Property, bool, error) {
	prop, ok, err := p.pf.FromAny(ctx, name, value, options...)
	if err != nil {
		return nil, false, err
	}

	if ok {
		p.AddProperty(ctx, prop)
	}
	return prop, ok, nil
}

// AddProperty adds the given property into the instance
func (p *Default) AddProperty(ctx context.Context, prop Property, options ...interface{}) error {
	p.syncMap.Store(prop.Name(ctx), prop)
	p.syncMapSize++
	return nil
}

// DeleteProperty removes the property
func (p *Default) DeleteProperty(ctx context.Context, prop Property, options ...interface{}) (bool, error) {
	return p.Delete(ctx, prop.Name(ctx), options...)
}

// Delete removes the property with the given name
func (p *Default) Delete(ctx context.Context, name PropertyName, options ...interface{}) (bool, error) {
	_, ok := p.syncMap.Load(name)
	if !ok {
		return false, nil
	}
	p.syncMap.Delete(name)
	p.syncMapSize--
	return true, nil
}

// Size returns the number of items in the list
func (p *Default) Size(context.Context) uint {
	return p.syncMapSize
}

// List returns all the properties as a slice
func (p *Default) List(context.Context) []Property {
	var result []Property
	p.syncMap.Range(func(key, value interface{}) bool {
		result = append(result, value.(Property))
		return true
	})
	return result
}

// Map returns all the properties as a map
func (p *Default) Map(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	p.syncMap.Range(func(key, value interface{}) bool {
		property := value.(Property)
		result[string(property.Name(ctx))] = property.AnyValue(ctx)
		return true
	})
	return result
}

// Named returns the named property and true if it was found, false if not
func (p *Default) Named(ctx context.Context, name PropertyName) (Property, bool) {
	prop, ok := p.syncMap.Load(name)
	if ok {
		return prop.(Property), true
	}
	return nil, false
}

// Filter returns the list of properties which match the filter criteria
func (p *Default) Filter(ctx context.Context, filter func(context.Context, Property) bool) []Property {
	var result []Property
	p.syncMap.Range(func(key, value interface{}) bool {
		property := value.(Property)
		if filter(ctx, property) {
			result = append(result, property)
		}
		return true
	})
	return result
}

// Range runs the do function on all entries
func (p *Default) Range(ctx context.Context, do func(context.Context, Property) bool) {
	p.syncMap.Range(func(key, value interface{}) bool {
		return do(ctx, value.(Property))
	})

}

// RangeNameValue runs the do function on all entries
func (p *Default) RangeNameValue(ctx context.Context, do func(context.Context, string, interface{})) {

}

func (p *Default) Write(context.Context, io.Writer, func(context.Context, io.Writer, Property) (bool, error)) error {
	panic("Not implemented yet!")
}
