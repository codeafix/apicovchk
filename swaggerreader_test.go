package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGetSwaggerContent(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/BarossaDataSwagger.json", strings.Replace(dir, "\\", "/", -1))
	sr, err := NewSwaggerReader(filepath)
	AssertSuccess(t, err)
	c, err := sr.GetSwaggerContent()
	AssertSuccess(t, err)
	AreEqual(t, "2.0", c.Swagger, "Incorrect Swagger deserialised")
}

func TestNewSwaggerReaderFailsWithInvalidURL(t *testing.T) {
	_, err := NewSwaggerReader("!£$%£%^$^&$%^")
	IsTrue(t, err != nil, "Did not fail creating a new SwaggerReader when passed an invalid URL")
}

func TestGetSwaggerContentReturnsError(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/DoesntExist.json", strings.Replace(dir, "\\", "/", -1))
	sr, err := NewSwaggerReader(filepath)
	AssertSuccess(t, err)
	_, err = sr.GetSwaggerContent()
	IsTrue(t, err != nil, "Error not returned")
}
