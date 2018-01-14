package main

//CovCheckerInfo holds a reference to the PathMap that contains all of the paths
//from the loaded Swagger files and will be used to track the Paths that have
//been used in a request from the transaction log files
type CovCheckerInfo struct {
	PathMap    *PathMap
	FileReader FileReader
}

//NewCovChecker returns a new instance of the Coverage Checker
func NewCovChecker() *CovCheckerInfo {
	return &CovCheckerInfo{
		PathMap: NewPathMap(),
	}
}

//CheckCoverage first reads all of the swagger files specified in the passed
//configuration. Then it loads all of the transaction log files specified and
//measures how much of the API in the swagger descriptions
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
