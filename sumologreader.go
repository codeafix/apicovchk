package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

//ParseSumoLogEntry creates a new RequestLogEntry from a line in the sumo log file
func (*SumoLogReaderInfo) ParseSumoLogEntry(logLine string) (RequestLogEntry, error) {
	rle := RequestLogEntry{}
	vals := strings.Split(logLine, ",")
	if len(vals) != 2 {
		return rle, fmt.Errorf("Not enough columns in line '%s'", logLine)
	}
	i := strings.Index(vals[0], " ")
	if i < 0 {
		return rle, fmt.Errorf("Incorrect format in first column '%s'", logLine)
	}
	url, err := url.Parse(vals[0][i+1:])
	if err != nil {
		return rle, err
	}
	els := strings.Split(url.Path, "/")
	if els[0] == "" {
		els = els[1:]
	}
	if len(els) < 1 || els[0] == "" {
		return rle, fmt.Errorf("Expecting 1 or more elements in URL path '%s'", url.Path)
	}
	path := "/" + strings.Join(els[1:], "/")
	rle.Method = strings.ToUpper(strings.TrimSpace(vals[0][0:i]))
	rle.Path = path
	rle.PathElements = els[1:]
	rle.URL = url
	rle.Query = url.Query()
	rle.Service = els[0]
	rle.Response = strings.TrimSpace(vals[1])
	return rle, nil
}

//SumoLogReaderInfo contains the URL the LogReader should read from
type SumoLogReaderInfo struct {
	URLReader URLReader
}

//LogReader is used to read data from a URL into an array of log entries
type LogReader interface {
	GetLogEntries() ([]RequestLogEntry, error)
}

//NewSumoLogReader returns a new instance of log reader
func NewSumoLogReader(urlstring string) (LogReader, error) {
	ur, err := NewURLReader(urlstring)
	if err != nil {
		return nil, err
	}
	return &SumoLogReaderInfo{URLReader: ur}, nil
}

//GetLogEntries reads the log entries from the transaction log's URL and returns a list of the
//parsed entries
func (slr *SumoLogReaderInfo) GetLogEntries() ([]RequestLogEntry, error) {
	c, err := slr.URLReader.ReadFromURL()
	if err != nil {
		return nil, err
	}
	rd := bufio.NewReader(bytes.NewReader(c))
	lel := []RequestLogEntry{}
	for {
		line, eof := rd.ReadString('\n')
		rle, err := slr.ParseSumoLogEntry(line)
		//Skip lines we can't parse
		if err == nil {
			lel = append(lel, rle)
		}
		if eof != nil {
			break
		}
	}
	return lel, nil
}
