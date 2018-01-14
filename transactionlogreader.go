package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const durationpos = 0
const startpos = 1
const endpos = 2
const methodpos = 3
const urlpos = 4
const bodypos = 5
const responsepos = 6

//TransactionLogEntry contains one entry from the transaction log that is produced by an application test framework.
//Any set of application tests that record the http requests in this format can be used to check the coverage of the
//API as defined in an associated Swagger description.
type TransactionLogEntry struct {
	Duration     int        `json:"duration"`
	Start        string     `json:"start"`
	End          string     `json:"end"`
	Method       string     `json:"method"`
	URL          *url.URL   `json:"url"`
	Path         string     `json:"path"`
	PathElements []string   `json:"pathElements"`
	Query        url.Values `json:"query"`
	Service      string     `json:"service"`
	Body         string     `json:"body"`
	Response     string     `json:"response"`
}

//ParseTransactionLogEntry creates a new TransactionLogEntry from a line in the TransactionLog file
func (*LogReaderInfo) ParseTransactionLogEntry(logLine string) (TransactionLogEntry, error) {
	tle := TransactionLogEntry{}
	vals := strings.Split(logLine, "\t")
	if len(vals) <= responsepos {
		return tle, fmt.Errorf("Not enough columns in line '%s'", logLine)
	}
	dur, err := strconv.Atoi(vals[durationpos])
	if err != nil {
		return tle, err
	}
	url, err := url.Parse(vals[urlpos])
	if err != nil {
		return tle, err
	}
	els := strings.Split(url.Path, "/")
	if els[0] == "" {
		els = els[1:]
	}
	if len(els) < 1 || els[0] == "" {
		return tle, fmt.Errorf("Expecting 1 or more elements in URL path '%s'", url.Path)
	}
	path := "/" + strings.Join(els[1:], "/")
	tle.Duration = dur
	tle.Start = strings.TrimSpace(vals[startpos])
	tle.End = strings.TrimSpace(vals[endpos])
	tle.Method = strings.ToUpper(strings.TrimSpace(vals[methodpos]))
	tle.URL = url
	tle.Path = path
	tle.PathElements = els[1:]
	tle.Query = url.Query()
	tle.Service = els[0]
	tle.Body = strings.TrimSpace(vals[bodypos])
	tle.Response = strings.TrimSpace(vals[responsepos])
	return tle, nil
}

//LogReaderInfo contains the URL the LogReader should read from
type LogReaderInfo struct {
	URLReader URLReader
}

//LogReader is used to read data from a URL into an array of log entries
type LogReader interface {
	GetLogEntries() ([]TransactionLogEntry, error)
}

//NewLogReader returns a new instance of log reader
func NewLogReader(urlstring string) (LogReader, error) {
	ur, err := NewURLReader(urlstring)
	if err != nil {
		return nil, err
	}
	return &LogReaderInfo{URLReader: ur}, nil
}

//GetLogEntries reads the log entries from the transaction log's URL and returns a list of the
//parsed entries
func (lr *LogReaderInfo) GetLogEntries() ([]TransactionLogEntry, error) {
	c, err := lr.URLReader.ReadFromURL()
	if err != nil {
		return nil, err
	}
	rd := bufio.NewReader(bytes.NewReader(c))
	lel := []TransactionLogEntry{}
	for {
		line, eof := rd.ReadString('\n')
		le, err := lr.ParseTransactionLogEntry(line)
		//Skip lines we can't parse
		if err == nil {
			lel = append(lel, le)
		}
		if eof != nil {
			break
		}
	}
	return lel, nil
}
