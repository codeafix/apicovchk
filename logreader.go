package main

import "net/url"

//LogReader is used to read data from a URL into an array of log entries
type LogReader interface {
	SetLogURL(urlstring string) error
	GetLogEntries() ([]RequestLogEntry, error)
}

//LogReaderInfo contains the URL the LogReader should read from
type LogReaderInfo struct {
	URLReader URLReader
}

//SetLogURL set's the URL to the log file that the reader should read
func (lr *LogReaderInfo) SetLogURL(urlstring string) error {
	ur, err := NewURLReader(urlstring)
	if err != nil {
		return err
	}
	lr.URLReader = ur
	return nil
}

//RequestLogEntry contains one entry from the request log that is produced by an application test framework.
//Any set of application tests that record the http requests in this format can be used to check the coverage of the
//API as defined in an associated Swagger description.
type RequestLogEntry struct {
	Method       string     `json:"method"`
	Path         string     `json:"path"`
	PathElements []string   `json:"pathElements"`
	URL          *url.URL   `json:"url"`
	Query        url.Values `json:"query"`
	Service      string     `json:"service"`
	Response     string     `json:"response"`
}
