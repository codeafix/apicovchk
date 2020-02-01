package main

import "net/url"

//RequestLogEntry contains one entry from the transaction log that is produced by an application test framework.
//Any set of application tests that record the http requests in this format can be used to check the coverage of the
//API as defined in an associated Swagger description.
type RequestLogEntry struct {
	Method       string     `json:"method"`
	Path         string     `json:"path"`
	PathElements []string   `json:"pathElements"`
	Query        url.Values `json:"query"`
	Service      string     `json:"service"`
	Response     string     `json:"response"`
}
