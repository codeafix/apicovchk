package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAddSwaggerToPathMap(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/BarossaDataSwagger.json", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "data",
				Swagger:   filepath,
			},
		},
	}
	pm := NewPathMap()
	err = pm.ReadSwagger(c)
	AssertSuccess(t, err)
	CheckGold(t, "PathMapFromSwaggerTest.json", pm.JSON())
}

func TestCheckTransactionLogEntry(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/coverage-report.txt", strings.Replace(dir, "\\", "/", -1))
	lr, err := NewLogReader(filepath)
	AssertSuccess(t, err)
	pm := NewPathMap()
	lel, err := lr.GetLogEntries()
	AssertSuccess(t, err)
	for _, entry := range lel {
		pm.CheckTransactionLogEntry(entry)
	}
	CheckGold(t, "PathMapFromLogTest.json", pm.JSON())
}
