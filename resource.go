package properties

import (
	"context"
	"github.com/lectio/resource"
	"github.com/spf13/afero"
	"net/url"
)

// URLProperty holds a URL
type URLProperty interface {
	Property
	URL(context.Context) *url.URL
}

// ResourceProperty holds a URL's resource
type ResourceProperty interface {
	Property
	Content(context.Context) resource.Content
}

// DownloadedResourceProperty holds a named file that was downloaded via an URL
type DownloadedResourceProperty interface {
	URLProperty
	LocalHRef(context.Context) string
	LocalFile(context.Context) (afero.Fs, string)
}

// DefaultResourceProperty implements ResourceProperty
type DefaultResourceProperty struct {
	PropName    PropertyName `json:"name"`
	ResourceURL *url.URL     `json:"url"`
	HREF        string       `json:"localHRef"`
	FilePath    afero.Fs     `json:"localFilePath"`
	FileName    string       `json:"localFileName"`
}

// Name returns the property name
func (p *DefaultResourceProperty) Name(context.Context) PropertyName {
	return p.PropName
}

// AnyValue returns the property value useful when the type isn't important
func (p *DefaultResourceProperty) AnyValue(context.Context) interface{} {
	return p.ResourceURL
}

// URL returns the associated URL
func (p *DefaultResourceProperty) URL(context.Context) *url.URL {
	return p.ResourceURL
}

// Content returns the page content and attachment
func (p *DefaultResourceProperty) Content(context.Context) resource.Content {
	panic("not implemented")
}

// LocalHRef returns the local href
func (p *DefaultResourceProperty) LocalHRef(context.Context) string {
	return p.HREF
}

// LocalFile returns the local file
func (p *DefaultResourceProperty) LocalFile(context.Context) (afero.Fs, string) {
	return p.FilePath, p.FileName
}
