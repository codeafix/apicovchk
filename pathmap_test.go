package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAddSwaggerToPathMap(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/PetstoreSwagger.json", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "petstore",
				Swagger:   filepath,
			},
		},
	}
	pm := NewPathMap()
	err = pm.ReadSwagger(c)
	AssertSuccess(t, err)
	CheckGold(t, "PathMapFromSwaggerTest.json", pm.JSON())
}

func CheckTransactionLogEntry(t *testing.T) {
	dir, err := os.Getwd()
	filepath := fmt.Sprintf("file:///%s/coverage-report.txt", strings.Replace(dir, "\\", "/", -1))
	lr := NewTransactionLogReader()
	err = lr.SetLogURL(filepath)
	AssertSuccess(t, err)
	pm := NewPathMap()
	lel, err := lr.GetLogEntries()
	AssertSuccess(t, err)
	for _, entry := range lel {
		pm.CheckRequestLogEntry(entry)
	}
	CheckGold(t, "PathMapFromLogTest.json", pm.JSON())
}

func GenerateAccessLogFromSwagger(t *testing.T) {
	dir, err := os.Getwd()
	AssertSuccess(t, err)
	entries := [][7]string{}
	sumo := true
	domain := "https://127.0.0.1:8081"
	if sumo {
		domain = ""
	}

	GenerateEntries(t, &entries, dir, domain)

	f, err := os.Create("temp/petstore-log.txt")
	AssertSuccess(t, err)
	defer f.Close()
	if sumo {
		f.WriteString("API,Response Code\n")
	} else {
		f.WriteString("duration(ms)\tstart-time\tend-time\tmethod\turl\tbody\tresponse\n")
	}
	for _, line := range entries {
		s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", line[0], line[1], line[2], line[3], line[4], line[5], line[6])
		if sumo {
			s = fmt.Sprintf("%s %s,%s\n", line[3], line[4], line[6])
		}
		f.WriteString(s)
	}
}

func GenerateEntries(t *testing.T, entries *[][7]string, dir string, domain string) {
	filepath := fmt.Sprintf("file:///%s/PetStoreSwagger.json", strings.Replace(dir, "\\", "/", -1))
	c := Config{
		Services: []ServiceEntry{
			ServiceEntry{
				RoutePath: "petstore",
				Swagger:   filepath,
			},
		},
	}
	pm := NewPathMap()
	err := pm.ReadSwagger(c)
	AssertSuccess(t, err)
	for _, pi := range pm.Services["petstore"].PathItems {
		WalkPathItems(entries, pi, fmt.Sprintf("%s/petstore", domain))
	}
}

func WalkPathItems(entries *[][7]string, pi *PathItem, path string) {

	key := pi.MapKey()
	if pi.IsParameter() {
		id, _ := uuid.NewUUID()
		key = id.String()
	}
	cpath := fmt.Sprintf("%s/%s", path, key)
	if pi.Verbs != nil {
		for _, verb := range pi.Verbs {
			if len(verb.Responses) > 0 {
				for _, response := range verb.Responses {
					entry := [7]string{
						"0",
						"0",
						"0",
						verb.Name,
						cpath,
						"undefined",
						response.Response,
					}
					*entries = append(*entries, entry)
				}
			} else {
				entry := [7]string{
					"0",
					"0",
					"0",
					verb.Name,
					cpath,
					"undefined",
					"200",
				}
				*entries = append(*entries, entry)
			}
			if len(verb.QueryParameters) > 0 {
				i := 0
				query := []string{}
				for _, param := range verb.QueryParameters {
					query = append(query, fmt.Sprintf("%s=val%d", param.Key, i))
					i++
				}
				cpath = fmt.Sprintf("%s?%s", cpath, strings.Join(query, "&"))
				entry := [7]string{
					"0",
					"0",
					"0",
					verb.Name,
					cpath,
					"undefined",
					"200",
				}
				*entries = append(*entries, entry)
			}
		}
	}
	for _, child := range pi.PathItems {
		WalkPathItems(entries, child, cpath)
	}
}
