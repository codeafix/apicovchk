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

//TransactionLogInfo contains the URLReader the LogReader should read from
type TransactionLogInfo struct {
	LogReaderInfo
}

//TransactionLogEntry contains one entry from the transaction log that is produced by an application test framework.
//Any set of application tests that record the http requests in this format can be used to check the coverage of the
//API as defined in an associated Swagger description.
type TransactionLogEntry struct {
	RequestLogEntry
	Duration int    `json:"duration"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Body     string `json:"body"`
}

//ParseTransactionLogEntry creates a new TransactionLogEntry from a line in the TransactionLog file
func (*TransactionLogInfo) ParseTransactionLogEntry(logLine string) (TransactionLogEntry, error) {
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

//NewTransactionLogReader returns a new instance of log reader
func NewTransactionLogReader() LogReader {
	return &TransactionLogInfo{}
}

//GetLogEntries reads the log entries from the transaction log's URL and returns a list of the
//parsed entries
func (lr *TransactionLogInfo) GetLogEntries() ([]RequestLogEntry, error) {
	c, err := lr.URLReader.ReadFromURL()
	if err != nil {
		return nil, err
	}
	rd := bufio.NewReader(bytes.NewReader(c))
	lel := []RequestLogEntry{}
	for {
		line, eof := rd.ReadString('\n')
		le, err := lr.ParseTransactionLogEntry(line)
		//Skip lines we can't parse
		if err == nil {
			lel = append(lel, le.RequestLogEntry)
		}
		if eof != nil {
			break
		}
	}
	return lel, nil
}
