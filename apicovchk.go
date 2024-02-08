package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	success, conf, outfilename := parseCommandLineOptions(os.Args)
	if success {
		err := covcheck(conf, outfilename)
		if err != nil {
			fmt.Printf("Error processing coverage check: %s\n\n", err.Error())
		}
	} else {
		printUsage()
	}
}

func covcheck(conf Config, outfilename string) error {
	cc := NewCovChecker()
	err := cc.CheckCoverage(conf)
	if err != nil {
		return err
	}
	cc.NavigatePathMap()
	hw := NewHTMLWriter(cc)
	return hw.Write(outfilename)
}

func parseCommandLineOptions(args []string) (success bool, conf Config, outfilename string) {
	success = true
	outfilename = "coverage.html"
	conf = Config{}
	if len(args) <= 1 || args[1] == "-help" {
		return false, conf, outfilename
	}
	skipnext := false
	for i, arg := range args[1:] {
		if skipnext {
			skipnext = false
			continue
		}
		switch arg {
		case "-opt":
			if len(args) < i+3 {
				fmt.Printf("<optionsFile> missing after -opt option \n\n")
				success = false
				break
			}
			skipnext = true
			file := args[i+2]
			optFile := readFile(file)
			err := json.Unmarshal([]byte(optFile), &conf)
			if err != nil {
				fmt.Printf("Error in <optionsFile>: %s \n\n", err.Error())
				success = false
				break
			}
		case "-out":
			if len(args) < i+3 {
				fmt.Printf("<covFileName> missing after -out option \n\n")
				success = false
				break
			}
			skipnext = true
			outfilename = args[i+2]
		default:
			fmt.Printf("Unrecognised option on command line: %s\n\n", args[i])
			success = false
		}
	}
	return success, conf, outfilename
}

func readFile(filename string) []byte {
	content, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}
	return content
}

func printUsage() {
	fmt.Println(`apicovchk command line utility

Takes an options file that lists a number of transaction logs and services each with
a Swagger 2.0 API description. This utility computes the coverage recorded in the
transaction files over an API as it is documented in the Swagger files.

Usage:
      apicovchk -opt <optionsFile> -out <covFileName>
      apicovchk -help

Options:
-opt <optionsFile>
      A file containing the list of transaction logs and the list of services in the API.
	  The file should be in the following format.
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
	  NOTE: Paths for swagger files and log files can be local file URLs or web URLs.
			There are two types of log file format supported:
			* Sumo: A comma separated file in the form:
					API,Response Code
					GET /petstore/pet/findByTags,200
					PUT /petstore/pet,200
					POST /petstore/pet,405
			* Transaction: A tab separated file in the form:
					duration(ms)	start-time	end-time	method	url	body	response
					663	18:55.0	18:55.6	POST	https://127.0.0.1:8081/petstore/user	undefined	200
					749	18:55.6	18:56.4	GET	https://127.0.0.1:8081/petstore/user/login	undefined	400
-out <covFileName>
      An HTML file containing the computed coverage report. If this option is not specified the utility
      create a file called "coverage.html" in the current directory.
-help
    Prints this message.`)
}
