package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseRequestLogEntry(t *testing.T) {
	ll := "GET /api/v3/Clients/*/Period/Document?year=*&filingMonth=*&productTypeId=*&productTypeId=*&productTypeId=*,200"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, "/api/v3/Clients/*/Period/Document?year=*&filingMonth=*&productTypeId=*&productTypeId=*&productTypeId=*", tle.URL.String(), "URL not correct")
	AreEqual(t, "/v3/Clients/*/Period/Document", tle.Path, "Path not correct")
	AreEqual(t, "api", tle.Service, "Service not correct")
	AreEqual(t, "200", tle.Response, "Response not correct")
}

func TestParseRequestLogEntryElems1(t *testing.T) {
	ll := "GET /api/v3/Clients/*/Period/Document?year=*&filingMonth=*&productTypeId=*&productTypeId=*&productTypeId=*,200"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 5, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "v3", tle.PathElements[0], "Path element 0 not correct")
	AreEqual(t, "Clients", tle.PathElements[1], "Path element 1 not correct")
	AreEqual(t, "*", tle.PathElements[2], "Path element 2 not correct")
	AreEqual(t, "Period", tle.PathElements[3], "Path element 3 not correct")
	AreEqual(t, "Document", tle.PathElements[4], "Path element 4 not correct")
}

func TestParseRequestLogEntryElems2(t *testing.T) {
	ll := "GET /api/v3/Clients/*/Period/Document/,200"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 6, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "v3", tle.PathElements[0], "Path element 0 not correct")
	AreEqual(t, "Clients", tle.PathElements[1], "Path element 1 not correct")
	AreEqual(t, "*", tle.PathElements[2], "Path element 2 not correct")
	AreEqual(t, "Period", tle.PathElements[3], "Path element 3 not correct")
	AreEqual(t, "Document", tle.PathElements[4], "Path element 4 not correct")
	AreEqual(t, "", tle.PathElements[5], "Path element 5 not correct")
}

func TestParseRequestLogEntryElems3(t *testing.T) {
	ll := "get http://127.0.0.1:58800/auth/?role=User&session=Active, 201"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 1, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "", tle.PathElements[0], "Path element 0 not correct")
}

func TestParseRequestLogEntryWorksWithFullPath(t *testing.T) {
	ll := "get http://127.0.0.1:58800/api/v3/Clients/*/Period/Document,200"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 5, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "v3", tle.PathElements[0], "Path element 0 not correct")
	AreEqual(t, "Clients", tle.PathElements[1], "Path element 1 not correct")
	AreEqual(t, "*", tle.PathElements[2], "Path element 2 not correct")
	AreEqual(t, "Period", tle.PathElements[3], "Path element 3 not correct")
	AreEqual(t, "Document", tle.PathElements[4], "Path element 4 not correct")
}

func TestParseRequestLogEntryFailsWithInvalidUrl(t *testing.T) {
	ll := "POST //--- - - orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods,404"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	IsTrue(t, err != nil, fmt.Sprintf("Incorrectly succeeded with invalid URL %v", tle.URL))
}

func TestParseRequestLogEntryFailsWithNoPathElementInUrl(t *testing.T) {
	ll := "POST http://127.0.0.1:58800/, 404"
	slr := &SumoLogReaderInfo{}
	_, err := slr.ParseSumoLogEntry(ll)
	IsTrue(t, err != nil, "Incorrectly succeeded with invalid URL")
}

func TestParseRequestLogEntryQuery(t *testing.T) {
	ll := "GET http://127.0.0.1:58800/auth/?role=User&session=Active,200"
	slr := &SumoLogReaderInfo{}
	tle, err := slr.ParseSumoLogEntry(ll)
	AssertSuccess(t, err)
	role, exists := tle.Query["role"]
	IsTrue(t, exists, "role query parameter does not exist")
	AreEqual(t, "User", role[0], "User not role parameter value")
	session, exists := tle.Query["session"]
	IsTrue(t, exists, "session query parameter does not exist")
	AreEqual(t, "Active", session[0], "Active not session parameter value")
}

func TestReadExampleSumoLog(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/sumologic.csv", strings.Replace(dir, "\\", "/", -1))

	lr := NewSumoLogReader()
	err = lr.SetLogURL(filepath)
	AssertSuccess(t, err)
	lel, err := lr.GetLogEntries()
	AssertSuccess(t, err)
	b, err := json.MarshalIndent(lel, "", "	")
	AssertSuccess(t, err)
	CheckGold(t, "LogEntriesFromSumoLogReader.json", string(b))
}
