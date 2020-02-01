package main

//Config contains the list of transaction log files to read, and an entry for each
//service that defines the reverse proxy path name for the service, and the
//swagger.json file to use.
type Config struct {
	TransactionLogs []string       `json:"transactionLogFiles"`
	Services        []ServiceEntry `json:"services"`
	LogFormat       int            `json:"logFormat"`
}

//ServiceEntry contains the path name used by the reverse proxy to route to the
//service and the URI of the swagger.json file to use for that service.
type ServiceEntry struct {
	RoutePath string `json:"routePath"`
	Swagger   string `json:"swagger"`
}
