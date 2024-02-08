# apicovchk [![Build Status](https://github.com/codeafix/apicovchk/actions/workflows/go_build.yml/badge.svg?branch=main)](https://github.com/codeafix/apicovchk/actions/workflows/go_build.yml) [![Coverage Status](http://codecov.io/github/codeafix/apicovchk/coverage.svg?branch=main)](http://codecov.io/github/codeafix/apicovchk?branch=main) [![BSD 3-Clause](https://img.shields.io/badge/License-BSD%203--Clause-green.svg)](https://github.com/codeafix/apicovchk/blob/master/LICENSE)
A simple utility for calculating how much of an API defined in a Swagger 2.0 API specification is exercised during a test. It works by computing coverage statistics of an API endpoint based on how many of the verbs, query parameters, and response codes that are defined in the swagger have actually been used from a log of all of the http requests in a test.

The computed coverage report is written out into an html file. Any request found in the specified http request log files increases the coverage statistic for that endpoint. Any endpoint definitions in the swagger file increase a documented statistic. For example, an endpoint that only appears in the http request logs will have a 100% coverage statistic, but a 0% documented statistic. Similarly and endpoint that only appears in the Swagger definition will have a 100% documented statistic, but a 0% coverage statistic.

The utility takes an options file that lists a number of transaction logs and services each with a Swagger 2.0 API description.

Usage:
```
    apicovchk -opt <optionsFile> -out <covFileName>
    apicovchk -help
```
Options:

`-opt <optionsFile>`
    A file containing the list of transaction logs and the list of services in the API. The file should be in the following format.
```
{
    "transactionLogFiles":[
        {
            "logURL": "log1.txt",
            "logType": "Sumo"
        },
        {
            "logURL": "file:///./logs/log2.txt",
            "logType": "Transaction"
        }
        ],
    "services":[
        {
            "routePath": "petstore",
            "swagger": "https://petstore.swagger.io/v2/swagger.json"
        },
        {
            "routePath": "open-api-spec",
            "swagger": "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/master/examples/v2.0/json/api-with-examples.json"
        }
    ]
}
```
NOTE: Paths for swagger files and log files can be local file URLs or web URLs.

There are two types of log file format supported.
* Sumo: A comma separated file in the form:
```
    API,Response Code
    GET /petstore/pet/findByTags,200
    PUT /petstore/pet,200
    POST /petstore/pet,405
```
* Transaction: A tab separated file in the form:
```
    duration(ms)	start-time	end-time	method	url	body	response
    663	18:55.0	18:55.6	POST	https://127.0.0.1:8081/petstore/user	undefined	200
    749	18:55.6	18:56.4	GET	https://127.0.0.1:8081/petstore/user/login	undefined	400
```

`-out <covFileName>`
    An HTML file containing the computed coverage report. If this option is not specified the utility create a file called "coverage.html" in the current directory.

`-help`
    Prints usage information.