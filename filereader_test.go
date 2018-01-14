package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
)

type TestFileReaderInfo struct {
	URL   *url.URL
	Error string
}

func NewTestFileReader(urlstring, errstr string) FileReader {
	url, err := url.Parse(urlstring)
	if err != nil {
		panic(err)
	}
	return &TestFileReaderInfo{
		URL:   url,
		Error: errstr,
	}
}

//URLScheme returns the scheme of the URL this FileReader has been created for
func (tfr *TestFileReaderInfo) URLScheme() string {
	return tfr.URL.Scheme
}

//URLString returns the full URL this FileReader has been created for
func (tfr *TestFileReaderInfo) URLString() string {
	return tfr.URL.String()
}

func (tfr *TestFileReaderInfo) ReadFromWeb() ([]byte, error) {
	if tfr.Error == "" {
		return []byte(`Read from web content`), nil
	}
	return nil, fmt.Errorf(tfr.Error)
}

func (tfr *TestFileReaderInfo) ReadFromFile() ([]byte, error) {
	if tfr.Error == "" {
		return []byte(`Read from file content`), nil
	}
	return nil, fmt.Errorf(tfr.Error)
}

func TestReadFromURLHttp(t *testing.T) {
	ur := &URLReaderInfo{FileReader: NewTestFileReader("http://www.domain.com/somefile.json", "")}
	c, err := ur.ReadFromURL()
	AssertSuccess(t, err)
	AreEqual(t, `Read from web content`, string(c), "Incorrect Swagger deserialised")
}

func TestReadFromURLHttps(t *testing.T) {
	ur := &URLReaderInfo{FileReader: NewTestFileReader("https://www.andomain.com/otherfile.json", "")}
	c, err := ur.ReadFromURL()
	AssertSuccess(t, err)
	AreEqual(t, `Read from web content`, string(c), "Incorrect Swagger deserialised")
}

func TestReadFromURLFile(t *testing.T) {
	ur := &URLReaderInfo{FileReader: NewTestFileReader("file:///tmp/otherfile.json", "")}
	c, err := ur.ReadFromURL()
	AssertSuccess(t, err)
	AreEqual(t, `Read from file content`, string(c), "Incorrect Swagger deserialised")
}

func TestReadFromUrlFailsWithUnrecognisedScheme(t *testing.T) {
	ur := &URLReaderInfo{FileReader: NewTestFileReader("zzz:///tmp/otherfile.json", "")}
	_, err := ur.ReadFromURL()
	IsTrue(t, err != nil, "Unrecognised URL Scheme did not fail")
}

func TestReadFromUrlFailsFailsWhenError(t *testing.T) {
	errstr := "Some error happened"
	ur := &URLReaderInfo{FileReader: NewTestFileReader("file:///tmp/otherfile.json", errstr)}
	_, err := ur.ReadFromURL()
	IsTrue(t, err != nil, "Read from file with error did not fail")
	AreEqual(t, errstr, err.Error(), "Incorrect message")
}

func TestNewFileReaderFailsWithInvalidURL(t *testing.T) {
	_, err := NewFileReader("!£$%£%^$^&$%^")
	IsTrue(t, err != nil, "Did not fail creating a new SwaggerReader when passed an invalid URL")
}

func TestReadFromWeb(t *testing.T) {
	sr, err := NewFileReader("https://www.google.com")
	AssertSuccess(t, err)
	c, err := sr.ReadFromWeb()
	AssertSuccess(t, err)
	AreEqual(t, "<!doctype html>", string(c[0:15]), "Unable to read google.com")
}

func TestReadFromFile(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/BarossaDataSwagger.json", strings.Replace(dir, "\\", "/", -1))
	sr, err := NewFileReader(filepath)
	AssertSuccess(t, err)
	c, err := sr.ReadFromFile()
	AreEqual(t, "\"swagger\": \"2.0\"", string(c[7:23]), "Unable to read file")
}
