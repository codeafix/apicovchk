package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCheckCoverage(t *testing.T) {
	dir, err := os.Getwd()
	swagpath := fmt.Sprintf("file:///%s/BarossaDataSwagger.json", strings.Replace(dir, "\\", "/", -1))
	logpath := fmt.Sprintf("file:///%s/coverage-report.txt", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "data",
				Swagger:   swagpath,
			},
		},
		TransactionLogs: []string{logpath},
	}
	cc := NewCovChecker()
	err = cc.CheckCoverage(c)
	AssertSuccess(t, err)

	CheckGold(t, "PathMapFromCovCheckTest.json", cc.PathMap.JSON())
}

func TestWriteOutput(t *testing.T) {
	dir, err := os.Getwd()
	swagpath := fmt.Sprintf("file:///%s/BarossaDataSwagger.json", strings.Replace(dir, "\\", "/", -1))
	logpath := fmt.Sprintf("file:///%s/coverage-report.txt", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "data",
				Swagger:   swagpath,
			},
		},
		TransactionLogs: []string{logpath},
	}
	cc := NewCovChecker()
	err = cc.CheckCoverage(c)
	AssertSuccess(t, err)
	cc.NavigatePathMap()
	f, err := os.Create("out.html")
	AssertSuccess(t, err)
	defer f.Close()
	_, err = f.Write([]byte(cc.PrintStats()))
}
