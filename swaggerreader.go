package main

import "github.com/go-openapi/spec"

//SwaggerReaderInfo contains the URL the SwaggerReader should read from
type SwaggerReaderInfo struct {
	URLReader URLReader
}

//SwaggerReader is used to read data from a URL into a github.com/go-openapi/spec.Swagger struct
type SwaggerReader interface {
	GetSwaggerContent() (*spec.Swagger, error)
}

//NewSwaggerReader returns a new instance of swagger reader
func NewSwaggerReader(urlstring string) (SwaggerReader, error) {
	ur, err := NewURLReader(urlstring)
	if err != nil {
		return nil, err
	}
	return &SwaggerReaderInfo{URLReader: ur}, nil
}

//GetSwaggerContent reads the Swagger from the SwaggerReader's URL and returns a spec.Swagger
func (sr *SwaggerReaderInfo) GetSwaggerContent() (*spec.Swagger, error) {
	swag := &spec.Swagger{}
	c, err := sr.URLReader.ReadFromURL()
	if err != nil {
		return swag, err
	}
	err = swag.UnmarshalJSON(c)
	return swag, err
}
