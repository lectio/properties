package properties

import (
	"context"
	"time"
)

// PropertyName is the name of a property
type PropertyName string

// A Property expresses a single front matter variable
type Property interface {
	Name(context.Context) PropertyName
	AnyValue(context.Context) interface{}
}

// TextProperty holds a named string
type TextProperty interface {
	Property
	Value(context.Context) string
}

// TextListProperty holds a named string slice
type TextListProperty interface {
	Property
	Value(context.Context) []string
}

// FlagProperty holds a named boolean flag
type FlagProperty interface {
	Property
	Value(context.Context) bool
}

// DateTimeProperty holds a named wall time
type DateTimeProperty interface {
	Property
	Value(context.Context) time.Time
}

// CardinalProperty holds a named cardinal value
type CardinalProperty interface {
	Property
	Value(context.Context) int64
}

// DefaultDateTimeProperty implements DateTimeProperty
type DefaultDateTimeProperty struct {
	PropName PropertyName `json:"name"`
	Time     time.Time    `json:"value"`
}

// Name returns the property name
func (p *DefaultDateTimeProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultDateTimeProperty) AnyValue(context.Context) interface{} {
	return p.Time
}

// Value returns the property value when the type is important
func (p *DefaultDateTimeProperty) Value(context.Context) time.Time {
	return p.Time
}

// DefaultFlagProperty implements FlagProperty
type DefaultFlagProperty struct {
	PropName PropertyName `json:"name"`
	Flag     bool         `json:"value"`
}

// Name returns the property name
func (p *DefaultFlagProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultFlagProperty) AnyValue(context.Context) interface{} {
	return p.Flag
}

// Value returns the property value when the type is important
func (p *DefaultFlagProperty) Value(context.Context) bool {
	return p.Flag
}

// DefaultCardinalProperty implements CardinalProperty
type DefaultCardinalProperty struct {
	PropName PropertyName `json:"name"`
	Number   int64        `json:"value"`
}

// Name returns the property name
func (p *DefaultCardinalProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultCardinalProperty) AnyValue(context.Context) interface{} {
	return p.Number
}

// Value returns the property value when the type is important
func (p *DefaultCardinalProperty) Value(context.Context) int64 {
	return p.Number
}

// DefaultTextProperty implements TextProperty
type DefaultTextProperty struct {
	PropName PropertyName `json:"name"`
	Text     string       `json:"value"`
}

// Name returns the property name
func (p *DefaultTextProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultTextProperty) AnyValue(context.Context) interface{} {
	return p.Text
}

// Value returns the property value when the type is important
func (p *DefaultTextProperty) Value(context.Context) string {
	return p.Text
}

// DefaultTextListProperty implements TextProperty
type DefaultTextListProperty struct {
	PropName PropertyName `json:"name"`
	Slice    []string     `json:"value"`
}

// Name returns the property name
func (p *DefaultTextListProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultTextListProperty) AnyValue(context.Context) interface{} {
	return p.Slice
}

// Value returns the property value when the type is important
func (p *DefaultTextListProperty) Value(context.Context) []string {
	return p.Slice
}
