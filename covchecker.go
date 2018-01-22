package main

import (
	"bytes"
	"fmt"
)

//CovCheckerInfo holds a reference to the PathMap that contains all of the paths
//from the loaded Swagger files and will be used to track the Paths that have
//been used in a request from the transaction log files
type CovCheckerInfo struct {
	PathMap      *PathMap
	FileReader   FileReader
	ServiceStats []*ServiceStat
}

//ServiceStat collects the aggregate coverage for an entire service
type ServiceStat struct {
	Name         string
	Endpoints    []EndpointStat
	Coverage     float64
	Undocumented float64
}

//EndpointStat collects the coverage stats for a specific endpoint
type EndpointStat struct {
	Path         string
	Verbs        []VerbStat
	Coverage     float64
	Undocumented float64
}

//VerbStat collects the coverage counts for a specific verb on
//an endpoint
type VerbStat struct {
	Method       string
	Total        int
	Covered      int
	Undocumented int
}

//NewCovChecker returns a new instance of the Coverage Checker
func NewCovChecker() *CovCheckerInfo {
	return &CovCheckerInfo{
		PathMap: NewPathMap(),
	}
}

//CheckCoverage first reads all of the swagger files specified in the passed
//configuration. Then it loads all of the transaction log files specified and
//measures how much of the API in the swagger descriptions has been accessed
//by the transactions in the log files
func (cc *CovCheckerInfo) CheckCoverage(config Config) error {
	err := cc.PathMap.ReadSwagger(config)
	if err != nil {
		return err
	}
	for _, logFile := range config.TransactionLogs {
		lr, err := NewLogReader(logFile)
		if err != nil {
			return err
		}
		lel, err := lr.GetLogEntries()
		if err != nil {
			return err
		}
		for _, entry := range lel {
			cc.PathMap.CheckTransactionLogEntry(entry)
		}
	}
	return nil
}

//PrintStats prints the calculated coverage stats into an HTML document
func (cc *CovCheckerInfo) PrintStats() string {
	var buf bytes.Buffer
	var totCov, totDoc float64
	for _, ss := range cc.ServiceStats {
		totCov += ss.Coverage
		totDoc += 1 - ss.Undocumented
	}
	totCov /= float64(len(cc.ServiceStats))
	totDoc /= float64(len(cc.ServiceStats))
	fmt.Fprintf(&buf, `<html>
<h1>Summary</h1>
<table>
<tr><td>Overall API coverage</td><td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> %3.2f%%</td></tr>
<tr><td>Overall API documentation</td><td><meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> %3.2f%%</td></tr>
</table>
<h2>Coverage summary by Service</h2>
<table>
`, totCov, totCov*100, totDoc, totDoc*100)

	//First let's print the summary of services
	for _, ss := range cc.ServiceStats {
		fmt.Fprintf(&buf,
			`<tr>
  <td>%s</td>
  <td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> Coverage: %3.2f%%<br/>
  <meter min="0"  max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> Documented: %3.2f%%
  </td>
</tr>
`, ss.Name, ss.Coverage, ss.Coverage*100, 1-ss.Undocumented, (1-ss.Undocumented)*100)
	}
	fmt.Fprintln(&buf, `</table>`)

	for _, ss := range cc.ServiceStats {
		fmt.Fprintf(&buf, `<h1>%s</h1>
<table>
`, ss.Name)

		for _, ep := range ss.Endpoints {
			fmt.Fprintf(&buf,
				`<tr>
  <td>%s</td>
  <td><meter min="0" max="1" low="0.8" high="0.8" optimum="1" value="%3.2f"></meter> Coverage: %3.2f%%<br/>
  <meter min="0" max="1" low=".9999" high=".9999" optimum="1" value="%3.2f"></meter> Documented: %3.2f%%
  </td>
</tr>
`, ep.Path, ep.Coverage, ep.Coverage*100, 1-ep.Undocumented, (1-ep.Undocumented)*100)
		}

		fmt.Fprintln(&buf, `</table>`)
	}
	fmt.Fprintln(&buf, `</html>`)
	return buf.String()
}

//NavigatePathMap navigates over the path map and calculates the coverage
//stats for each verb on the path, overall stats for the path, and for
//the services
func (cc *CovCheckerInfo) NavigatePathMap() {
	cc.ServiceStats = []*ServiceStat{}
	for sn, srv := range cc.PathMap.Services {
		ss := &ServiceStat{
			Name:      sn,
			Endpoints: []EndpointStat{},
		}
		cc.ServiceStats = append(cc.ServiceStats, ss)
		cc.NavigatePathItem(ss, srv, "")
		ss.Coverage = ss.Coverage / float64(len(ss.Endpoints))
		ss.Undocumented = ss.Undocumented / float64(len(ss.Endpoints))
	}
}

//NavigatePathItem iterates over path items descending the path hierarchy
//and uses another function to generate the stats for each verb it finds
func (cc *CovCheckerInfo) NavigatePathItem(ss *ServiceStat, pi *PathItem, path string) {
	for _, child := range pi.PathItems {
		cpath := fmt.Sprintf("%s/%s", path, child.MapKey())
		var tot, cov, und int
		if child.Verbs != nil {
			es := EndpointStat{
				Path:  cpath,
				Verbs: []VerbStat{},
			}
			for _, verb := range child.Verbs {
				vs := cc.CalculateVerbStats(verb)
				tot = tot + vs.Total
				cov = cov + vs.Covered
				und = und + vs.Undocumented
				es.Verbs = append(es.Verbs, vs)
			}
			es.Coverage = float64(cov) / float64(tot)
			es.Undocumented = float64(und) / float64(tot)
			ss.Endpoints = append(ss.Endpoints, es)
			ss.Coverage = ss.Coverage + es.Coverage
			ss.Undocumented = ss.Undocumented + es.Undocumented
		}
		cc.NavigatePathItem(ss, child, cpath)
	}
}

//CalculateVerbStats generates a coverage statistic for the verb by dividing
//the logged response codes and query parameters against the documented
//response codes and query parameters
func (cc *CovCheckerInfo) CalculateVerbStats(verb *Verb) VerbStat {
	vs := VerbStat{Method: verb.Name}
	for _, response := range verb.Responses {
		vs.Total = vs.Total + 1
		if !response.Documented {
			vs.Undocumented = vs.Undocumented + 1
		}
		if response.Covered > 0 {
			vs.Covered = vs.Covered + 1
		}
	}
	for _, param := range verb.QueryParameters {
		vs.Total = vs.Total + 1
		if !param.Documented {
			vs.Undocumented = vs.Undocumented + 1
		}
		if param.Covered > 0 {
			vs.Covered = vs.Covered + 1
		}
	}
	return vs
}
