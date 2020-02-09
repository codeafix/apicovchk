package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCheckCoverage(t *testing.T) {
	dir, err := os.Getwd()
	swagpath := fmt.Sprintf("file:///%s/PetstoreSwagger.json", strings.Replace(dir, "\\", "/", -1))
	logpath := fmt.Sprintf("file:///%s/petstore-report.txt", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "petstore",
				Swagger:   swagpath,
			},
		},
		TransactionLogs: []LogEntry{
			LogEntry{
				LogURL:  logpath,
				LogType: Transaction,
			},
		},
	}
	cc := NewCovChecker()
	err = cc.CheckCoverage(c)
	AssertSuccess(t, err)

	CheckGold(t, "PathMapFromCovCheckTest.json", cc.PathMap.JSON())
}

func TestWriteOutput(t *testing.T) {
	dir, err := os.Getwd()
	swagpath := fmt.Sprintf("file:///%s/PetstoreSwagger.json", strings.Replace(dir, "\\", "/", -1))
	logpath := fmt.Sprintf("file:///%s/petstore-report.txt", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "petstore",
				Swagger:   swagpath,
			},
		},
		TransactionLogs: []LogEntry{
			LogEntry{
				LogURL:  logpath,
				LogType: Transaction,
			},
		},
	}
	cc := NewCovChecker()
	err = cc.CheckCoverage(c)
	AssertSuccess(t, err)
	cc.NavigatePathMap()
	hw := NewHTMLWriter(cc)
	err = hw.Write("temp/out.html")
	AssertSuccess(t, err)
}
