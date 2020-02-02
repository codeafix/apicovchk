package main

import (
	"errors"
	"strings"
)

//Config contains the list of transaction log files to read, and an entry for each
//service that defines the reverse proxy path name for the service, and the
//swagger.json file to use.
type Config struct {
	TransactionLogs []LogEntry     `json:"transactionLogFiles"`
	Services        []ServiceEntry `json:"services"`
}

//ServiceEntry contains the path name used by the reverse proxy to route to the
//service and the URI of the swagger.json file to use for that service.
type ServiceEntry struct {
	RoutePath string `json:"routePath"`
	Swagger   string `json:"swagger"`
}

//LogEntry contains the path and type of a log file to read
type LogEntry struct {
	LogURL  string  `json:"logURL"`
	LogType LogType `json:"logType"`
}

//LogType is an enum to indicate the format of the log file to be read
type LogType string

const (
	//Sumo the format of log file that is exported from request logs
	Sumo = "Sumo"
	//Transaction the format of log file exported directly from test code
	Transaction = "Transaction"
)

//UnmarshalJSON implements parsing of a string representation during json deserialise into LogType
func (lt *LogType) UnmarshalJSON(b []byte) error {
	logType := LogType(strings.Trim(string(b), `"`))
	switch logType {
	case Sumo, Transaction:
		*lt = logType
		return nil
	}
	return errors.New("Invalid log type")
}
