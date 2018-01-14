package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParseCommandOptionsFailsWhenOnlyHelp(t *testing.T) {
	args := []string{"apicovchk.exe", "-help"}
	success, _, _ := parseCommandLineOptions(args)
	IsFalse(t, success, "Expected parse to fail")
}

func TestParseCommandOptionsReadOptionsFile(t *testing.T) {
	args := []string{"apicovchk.exe", "-opt", "options.json"}
	success, conf, outfilename := parseCommandLineOptions(args)
	IsTrue(t, success, "Expected parse to succeed")
	AreEqual(t, "coverage.html", outfilename, "outfilename not set to default")
	AreEqual(t, 2, len(conf.TransactionLogs), "Wrong number of transaction logs")
	AreEqual(t, "log1.txt", conf.TransactionLogs[0], "Wrong transaction log filename")
	AreEqual(t, "log2.txt", conf.TransactionLogs[1], "Wrong transaction log filename")
	AreEqual(t, 2, len(conf.Services), "Wrong number of services")
	AreEqual(t, "data", conf.Services[0].RoutePath, "Wrong service route name")
	AreEqual(t, "report-image-generation", conf.Services[1].RoutePath, "Wrong service route name")
	AreEqual(t, "http://barossa-data.ci-iso.barossabuild.com:6100/Swagger/BarossaDataSwagger.json", conf.Services[0].Swagger, "Wrong swagger URL")
	AreEqual(t, "http://report-image-gen-svc.ci.oap.int.thomsonreuters.com:5110/Swagger/ReportImageGenerationSwagger.json", conf.Services[1].Swagger, "Wrong swagger URL")
}

func TestParseCommandOptionsFailsWhenNoOptfile(t *testing.T) {
	args := []string{"apicovchk.exe", "-opt"}
	success, _, _ := parseCommandLineOptions(args)
	IsFalse(t, success, "Expected parse to fail")
}

func TestParseCommandOptionsFailsWhenNotValidJson(t *testing.T) {
	args := []string{"apicovchk.exe", "-opt", "coverage-report.txt"}
	success, _, _ := parseCommandLineOptions(args)
	IsFalse(t, success, "Expected parse to fail")
}

func TestParseCommandOptionsFailsWhenOptfileDoesntExist(t *testing.T) {
	args := []string{"apicovchk.exe", "-opt", "doesntExist.json"}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	parseCommandLineOptions(args)
}

func TestParseCommandOptionsSetsOutFile(t *testing.T) {
	args := []string{"apicovchk.exe", "-out", "myCovOut.html"}
	success, _, outfilename := parseCommandLineOptions(args)
	IsTrue(t, success, "Expected parse to succeed")
	AreEqual(t, "myCovOut.html", outfilename, "outfilename wrong")
}

func TestParseCommandOptionsFailsWhenNoOutfile(t *testing.T) {
	args := []string{"apicovchk.exe", "-out"}
	success, _, _ := parseCommandLineOptions(args)
	IsFalse(t, success, "Expected parse to fail")
}

func TestParseCommandOptionsFailsWithUnrecognisedOption(t *testing.T) {
	args := []string{"apicovchk.exe", "-sss", "myCovOut.html"}
	success, _, _ := parseCommandLineOptions(args)
	IsFalse(t, success, "Expected parse to fail")
}

func IsFalse(t *testing.T, condition bool, msg string) {
	if condition {
		t.Error(msg)
	}
}

func IsTrue(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}

func AreEqual(t *testing.T, expected interface{}, actual interface{}, msg string) {
	if expected != actual {
		t.Errorf("%s Expected = %v Actual = %v", msg, expected, actual)
	}
}

func NotEqual(t *testing.T, expected interface{}, actual interface{}, msg string) {
	if expected == actual {
		t.Errorf("%s Expected = %v Actual = %v", msg, expected, actual)
	}
}

func AssertSuccess(t *testing.T, err error) {
	if err != nil {
		t.Errorf(err.Error())
	}
}

func CheckGold(t *testing.T, goldfile, json string) {
	gold, err := ioutil.ReadFile(goldfile)
	AssertSuccess(t, err)

	if string(gold) != json {
		f, err := os.Create("check_" + goldfile)
		AssertSuccess(t, err)
		defer f.Close()
		_, err = f.Write([]byte(json))
		AssertSuccess(t, err)
		t.Errorf("The generated path map is different from the gold file")
	}
}
