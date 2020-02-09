package main

import (
	"fmt"
)

//CovCheckerInfo holds a reference to the PathMap that contains all of the paths
//from the loaded Swagger files and will be used to track the Paths that have
//been used in a request from the transaction log files
type CovCheckerInfo struct {
	PathMap      *PathMap
	FileReader   FileReader
	ServiceStats []*ServiceStat
	Coverage     float64
	Undocumented float64
	TotalPoint   float64
}

//ServiceStat collects the aggregate coverage for an entire service
type ServiceStat struct {
	Name         string
	Endpoints    []EndpointStat
	Coverage     float64
	Undocumented float64
	TotalPoint   float64
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
	Responses    map[string]*Response
	Parameters   map[string]*QueryParameter
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
		lrr := NewLogReaderRepo()
		lr := lrr.GetLogReader(logFile.LogType)
		err := lr.SetLogURL(logFile.LogURL)
		if err != nil {
			return err
		}
		lel, err := lr.GetLogEntries()
		if err != nil {
			return err
		}
		for _, entry := range lel {
			cc.PathMap.CheckRequestLogEntry(entry)
		}
	}
	return nil
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
		ss.Coverage = ss.Coverage / ss.TotalPoint
		ss.Undocumented = ss.Undocumented / ss.TotalPoint
	}
	cc.Coverage = cc.Coverage / cc.TotalPoint
	cc.Undocumented = cc.Undocumented / cc.TotalPoint
}

//NavigatePathItem iterates over path items descending the path hierarchy
//and uses another function to generate the stats for each verb it finds
func (cc *CovCheckerInfo) NavigatePathItem(ss *ServiceStat, pi *PathItem, path string) {
	for _, child := range pi.PathItems {
		cpath := fmt.Sprintf("%s/%s", path, child.MapKey())
		var tot, cov, und float64
		if child.Verbs != nil {
			es := EndpointStat{
				Path:  cpath,
				Verbs: []VerbStat{},
			}
			for _, verb := range child.Verbs {
				vs := cc.CalculateVerbStats(verb)
				tot = tot + float64(vs.Total)
				cov = cov + float64(vs.Covered)
				und = und + float64(vs.Undocumented)
				es.Verbs = append(es.Verbs, vs)
			}
			es.Coverage = cov / tot
			es.Undocumented = und / tot
			ss.Endpoints = append(ss.Endpoints, es)
			ss.Coverage += cov
			ss.Undocumented += und
			ss.TotalPoint += tot
			cc.Coverage += cov
			cc.Undocumented += und
			cc.TotalPoint += tot
		}
		cc.NavigatePathItem(ss, child, cpath)
	}
}

//CalculateVerbStats generates a coverage statistic for the verb by dividing
//the logged response codes and query parameters against the documented
//response codes and query parameters
func (cc *CovCheckerInfo) CalculateVerbStats(verb *Verb) VerbStat {
	vs := VerbStat{
		Method:     verb.Name,
		Responses:  verb.Responses,
		Parameters: verb.QueryParameters,
	}
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
