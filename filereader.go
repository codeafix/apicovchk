package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

//FileReaderInfo contains the URL the FileReader should read from
type FileReaderInfo struct {
	URL *url.URL
}

//FileReader is used to read data from a URL into a byte array
type FileReader interface {
	ReadFromWeb() ([]byte, error)
	ReadFromFile() ([]byte, error)
	URLScheme() string
	URLString() string
}

//URLReaderInfo contains the a FileReader to read the File
type URLReaderInfo struct {
	FileReader
}

//URLReader is used to read data from a URL into a byte array
//This code is separated into a composite interface so that the code used to
//decide whether to read from a web address or from local storage can be
//tested without having to actually read from a web address or local disk
type URLReader interface {
	FileReader
	ReadFromURL() ([]byte, error)
}

//NewURLReader returns a new instance of a URL reader
func NewURLReader(urlstring string) (URLReader, error) {
	fr, err := NewFileReader(urlstring)
	if err != nil {
		return nil, err
	}
	return &URLReaderInfo{FileReader: fr}, nil
}

//NewFileReader returns a new instance of a file reader
func NewFileReader(urlstring string) (FileReader, error) {
	url, err := url.Parse(urlstring)
	if err != nil {
		return nil, fmt.Errorf("Invalid URL for file '%s': %s", urlstring, err)
	}
	return &FileReaderInfo{URL: url}, nil
}

//URLScheme returns the scheme of the URL this FileReader has been created for
func (fr *FileReaderInfo) URLScheme() string {
	return fr.URL.Scheme
}

//URLString returns the full URL this FileReader has been created for
func (fr *FileReaderInfo) URLString() string {
	return fr.URL.String()
}

//ReadFromWeb reads a file from a web URL
func (fr *FileReaderInfo) ReadFromWeb() ([]byte, error) {
	req, err := http.NewRequest("GET", fr.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

//ReadFromFile reads a file from local storage
func (fr *FileReaderInfo) ReadFromFile() ([]byte, error) {
	filepath := fr.URL.Path
	if filepath[0] == '/' {
		filepath = filepath[1:]
	}
	return ioutil.ReadFile(filepath)
}

//ReadFromURL reads the content from the URL specified in this FileReader
func (ur *URLReaderInfo) ReadFromURL() ([]byte, error) {
	var c []byte
	var err error
	switch ur.URLScheme() {
	case "http", "https":
		c, err = ur.ReadFromWeb()
	case "file":
		c, err = ur.ReadFromFile()
	default:
		return nil, fmt.Errorf("Unsupported scheme in log URL: '%s'", ur.URLString())
	}
	return c, err
}
