package properties

import (
	"context"
	"fmt"
	"sync"
)

// AddPropertyPolicy can prevent a property from being added
type AddPropertyPolicy interface {
	AllowAdd(context.Context, Property, ...interface{}) (Property, bool, error)
}

// AddPropertyEvent announces when a property has been added
type AddPropertyEvent interface {
	PropertyAdded(context.Context, Property, ...interface{})
}

// MapAssignFunc is passed into Properties.Map() to assign values into a string map
type MapAssignFunc func(context.Context, Property, map[string]interface{}) bool

// Properties manages a group of strongly typed properties, immutable
type Properties interface {
	List(context.Context, ...interface{}) []Property
	Map(context.Context, map[string]interface{}, MapAssignFunc, ...interface{}) uint
	Named(context.Context, PropertyName) (Property, bool)
	Filter(context.Context, func(context.Context, Property) bool, ...interface{}) []Property
	Range(context.Context, func(context.Context, Property) bool, ...interface{})
	Size(context.Context) uint
}

// AllowAddFunc returns true if the property should be added
type AllowAddFunc func(context.Context, string, interface{}, Property) (Property, bool, error)

// AllowAddTextFunc returns true if the property should be added
type AllowAddTextFunc func(context.Context, string, string, Property) (Property, bool, error)

// MutableProperties adds mutability to Properties
type MutableProperties interface {
	Properties
	AddMap(context.Context, map[string]interface{}, AllowAddFunc, ...interface{}) (uint, error)
	AddTextMap(context.Context, map[string]string, AllowAddTextFunc, ...interface{}) (uint, error)
	AddChecked(context.Context, string, interface{}, AllowAddFunc, ...interface{}) (Property, bool, error)
	AddParsedChecked(context.Context, string, string, AllowAddTextFunc, ...interface{}) (Property, bool, error)
	Add(context.Context, string, interface{}, ...interface{}) (Property, bool, error)
	AddParsed(context.Context, string, string, ...interface{}) (Property, bool, error)
	AddProperty(context.Context, Property, ...interface{}) (Property, bool, error)
	Delete(context.Context, PropertyName, ...interface{}) (bool, error)
	DeleteProperty(context.Context, Property, ...interface{}) (bool, error)
}

// Default is the default properties implementation (supports mutability)
type Default struct {
	pf          PropertyFactory
	syncMap     sync.Map
	syncMapSize uint
	addPolicy   AddPropertyPolicy
	addEvent    AddPropertyEvent
}

func newDefaultProperties(ctx context.Context, pf PropertyFactory, options ...interface{}) *Default {
	result := &Default{pf: pf}

	for _, option := range options {
		if instance, ok := option.(AddPropertyPolicy); ok {
			result.addPolicy = instance
		}
		if instance, ok := option.(AddPropertyEvent); ok {
			result.addEvent = instance
		}
	}

	return result
}

// DefaultAllowAdd is passed into AddMap returns true if the property should be added
func DefaultAllowAdd(ctx context.Context, givenName string, givenValue interface{}, createdProp Property) (Property, bool, error) {
	return createdProp, true, nil
}

// AddMap adds all the items in the given map
func (p *Default) AddMap(ctx context.Context, items map[string]interface{}, allow AllowAddFunc, options ...interface{}) (uint, error) {
	if items == nil {
		return 0, fmt.Errorf("items is Nil in properties.Default.AddMap")
	}

	var count uint
	for name, value := range items {
		_, ok, err := p.AddChecked(ctx, name, value, allow, options...)
		if err != nil {
			return count, err
		}
		if ok {
			count++
		}
	}

	return count, nil
}

// DefaultAllowAddTextFunc returns true if the property should be added
func DefaultAllowAddTextFunc(ctx context.Context, givenName string, givenValue string, createdProp Property) (Property, bool, error) {
	return createdProp, true, nil
}

// AddTextMap adds all the items in the given map by trying to "smart parse" the text
func (p *Default) AddTextMap(ctx context.Context, items map[string]string, allow AllowAddTextFunc, options ...interface{}) (uint, error) {
	if items == nil {
		return 0, fmt.Errorf("items is Nil in properties.Default.AddTextMap")
	}

	var count uint
	for name, value := range items {
		_, ok, err := p.AddParsedChecked(ctx, name, value, allow, options...)
		if err != nil {
			return count, err
		}
		if ok {
			count++
		}
	}

	return count, nil
}

// AddParsedChecked adds a single named property of a text value by "smart parsing" the value type
func (p *Default) AddParsedChecked(ctx context.Context, name string, value string, allow AllowAddTextFunc, options ...interface{}) (Property, bool, error) {
	prop, ok, err := p.pf.FromText(ctx, name, value, options...)
	if err != nil {
		return nil, false, err
	}

	if allow != nil {
		prop, ok, err = allow(ctx, name, value, prop)
	}

	if ok {
		return p.AddProperty(ctx, prop)
	}
	return prop, ok, nil
}

// AddChecked adds a single named property of any value type
func (p *Default) AddChecked(ctx context.Context, name string, value interface{}, allow AllowAddFunc, options ...interface{}) (Property, bool, error) {
	prop, ok, err := p.pf.FromAny(ctx, name, value, options...)
	if err != nil {
		return nil, false, err
	}

	if allow != nil {
		prop, ok, err = allow(ctx, name, value, prop)
	}

	if ok {
		return p.AddProperty(ctx, prop)
	}
	return prop, ok, nil
}

// AddParsed adds a single named property of a text value by "smart parsing" the value type
func (p *Default) AddParsed(ctx context.Context, name string, value string, options ...interface{}) (Property, bool, error) {
	return p.AddParsedChecked(ctx, name, value, nil, options...)
}

// Add adds a single named property of any value type
func (p *Default) Add(ctx context.Context, name string, value interface{}, options ...interface{}) (Property, bool, error) {
	return p.AddChecked(ctx, name, value, nil, options...)
}

// AddProperty adds the given property into the instance
func (p *Default) AddProperty(ctx context.Context, givenProp Property, options ...interface{}) (Property, bool, error) {
	finalProp := givenProp
	if p.addPolicy != nil {
		var add bool
		var err error
		finalProp, add, err = p.addPolicy.AllowAdd(ctx, givenProp, options...)
		if err != nil {
			return givenProp, false, err
		}
		if !add {
			return finalProp, false, nil
		}
	}

	p.syncMap.Store(finalProp.Name(ctx), finalProp)
	p.syncMapSize++

	if p.addEvent != nil {
		p.addEvent.PropertyAdded(ctx, finalProp, options...)
	}

	return finalProp, true, nil
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
func (p *Default) List(context.Context, ...interface{}) []Property {
	var result []Property
	p.syncMap.Range(func(key, value interface{}) bool {
		result = append(result, value.(Property))
		return true
	})
	return result
}

// DefaultMapAssign is passed into Map() for default property assignment rule
func DefaultMapAssign(ctx context.Context, p Property, dest map[string]interface{}) bool {
	p.Copy(ctx, dest)
	return true
}

// Map returns all the properties as a map
func (p *Default) Map(ctx context.Context, dest map[string]interface{}, assign MapAssignFunc, options ...interface{}) uint {
	var count uint
	p.syncMap.Range(func(key, value interface{}) bool {
		property := value.(Property)
		keepGoing := assign(ctx, property, dest)
		if keepGoing {
			count++
		}
		return keepGoing
	})
	return count
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
func (p *Default) Filter(ctx context.Context, filter func(context.Context, Property) bool, options ...interface{}) []Property {
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
func (p *Default) Range(ctx context.Context, do func(context.Context, Property) bool, options ...interface{}) {
	p.syncMap.Range(func(key, value interface{}) bool {
		return do(ctx, value.(Property))
	})
}
