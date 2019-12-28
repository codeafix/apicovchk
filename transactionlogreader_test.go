package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseTransactionLogEntry(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	POST	http://127.0.0.1:58800/orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods	\"{\"\"startDate\"\":\"\"2016-12-31\"\",\"\"endDate\"\":\"\"2017-12-30\"\",\"\"publishedId\"\":\"\"0af343ae-4468-44e5-98aa-897e6e6c5458\"\"}\"	201"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, 1828, tle.Duration, "Duration not correct")
	AreEqual(t, "18:57.8", tle.Start, "Start not correct")
	AreEqual(t, "18:59.7", tle.End, "End not correct")
	AreEqual(t, "POST", tle.Method, "Method not correct")
	AreEqual(t, "http://127.0.0.1:58800/orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods", tle.URL.String(), "URL not correct")
	AreEqual(t, "/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods", tle.Path, "Path not correct")
	AreEqual(t, "orchestration", tle.Service, "Service not correct")
	AreEqual(t, "\"{\"\"startDate\"\":\"\"2016-12-31\"\",\"\"endDate\"\":\"\"2017-12-30\"\",\"\"publishedId\"\":\"\"0af343ae-4468-44e5-98aa-897e6e6c5458\"\"}\"", tle.Body, "Body not correct")
	AreEqual(t, "201", tle.Response, "Response not correct")
}

func TestParseTransactionLogEntryElems1(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	get	http://127.0.0.1:58800/orchestration/client/TestClient/entity	undefined	201"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 3, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "client", tle.PathElements[0], "Path element 0 not correct")
	AreEqual(t, "TestClient", tle.PathElements[1], "Path element 1 not correct")
	AreEqual(t, "entity", tle.PathElements[2], "Path element 2 not correct")
}

func TestParseTransactionLogEntryElems2(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	get	http://127.0.0.1:58800/orchestration/client/TestClient/entity/	undefined	201"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 4, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "client", tle.PathElements[0], "Path element 0 not correct")
	AreEqual(t, "TestClient", tle.PathElements[1], "Path element 1 not correct")
	AreEqual(t, "entity", tle.PathElements[2], "Path element 2 not correct")
	AreEqual(t, "", tle.PathElements[3], "Path element 3 not correct")
}

func TestParseTransactionLogEntryElems3(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	get	http://127.0.0.1:58800/auth/?role=User&session=Active	undefined	201"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "GET", tle.Method, "Method not correct")
	AreEqual(t, 1, len(tle.PathElements), "Wrong number of path elements")
	AreEqual(t, "", tle.PathElements[0], "Path element 0 not correct")
}

func TestParseTransactionLogEntryWorksWithRelativePath(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	POST	orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods	\"{\"\"startDate\"\":\"\"2016-12-31\"\",\"\"endDate\"\":\"\"2017-12-30\"\",\"\"publishedId\"\":\"\"0af343ae-4468-44e5-98aa-897e6e6c5458\"\"}\"	404"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	AreEqual(t, "orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods", tle.URL.String(), "URL not correct")
	AreEqual(t, "/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods", tle.Path, "Path not correct")
	AreEqual(t, "orchestration", tle.Service, "Service not correct")
}

func TestParseTransactionLogEntryReportsFailsWhenDurationNaN(t *testing.T) {
	ll := "whatever	18:57.8	18:59.7	POST	orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods	\"{\"\"startDate\"\":\"\"2016-12-31\"\",\"\"endDate\"\":\"\"2017-12-30\"\",\"\"publishedId\"\":\"\"0af343ae-4468-44e5-98aa-897e6e6c5458\"\"}\"	404"
	tlr := &LogReaderInfo{}
	_, err := tlr.ParseTransactionLogEntry(ll)
	IsTrue(t, err != nil, "Incorrectly succeeded when duration NaN")
}

func TestParseTransactionLogEntryFailsWithInvalidUrl(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	POST	//--- - - orchestration/client/TestClient/entity/32a7e0b0-8130-4ab1-ace0-a81000890a14/periods	\"{\"\"startDate\"\":\"\"2016-12-31\"\",\"\"endDate\"\":\"\"2017-12-30\"\",\"\"publishedId\"\":\"\"0af343ae-4468-44e5-98aa-897e6e6c5458\"\"}\"	404"
	tlr := &LogReaderInfo{}
	_, err := tlr.ParseTransactionLogEntry(ll)
	IsTrue(t, err != nil, "Incorrectly succeeded with invalid URL")
}

func TestParseTransactionLogEntryFailsWithNoPathElementInUrl(t *testing.T) {
	ll := "1828	18:57.8	18:59.7	POST	http://127.0.0.1:58800/	undefined	404"
	tlr := &LogReaderInfo{}
	_, err := tlr.ParseTransactionLogEntry(ll)
	IsTrue(t, err != nil, "Incorrectly succeeded with invalid URL")
}

func TestParseTransactionLogEntryQuery(t *testing.T) {
	ll := "663	18:55.0	18:55.6	GET	http://127.0.0.1:58800/auth/?role=User&session=Active	undefined	200"
	tlr := &LogReaderInfo{}
	tle, err := tlr.ParseTransactionLogEntry(ll)
	AssertSuccess(t, err)
	role, exists := tle.Query["role"]
	IsTrue(t, exists, "role query parameter does not exist")
	AreEqual(t, "User", role[0], "User not role parameter value")
	session, exists := tle.Query["session"]
	IsTrue(t, exists, "session query parameter does not exist")
	AreEqual(t, "Active", session[0], "Active not session parameter value")
}

func TestReadExampleLog(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/coverage-report.txt", strings.Replace(dir, "\\", "/", -1))

	lr, err := NewLogReader(filepath)
	AssertSuccess(t, err)
	lel, err := lr.GetLogEntries()
	AssertSuccess(t, err)
	b, err := json.MarshalIndent(lel, "", "	")
	AssertSuccess(t, err)
	CheckGold(t, "LogEntriesFromLogReader.json", string(b))
}
