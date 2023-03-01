package main

import (
	"goris/client"
	"goris/common"
	"os"
	"strconv"
)

func main() {

	args := os.Args

	var port int
	var requests int
	var err error = nil

	// First arg is a reference to the binary itself.
	if len(args) == 1 {
		requests = 1
		port = common.DEF_SERVER_PORT
	} else {
		requests, err = strconv.Atoi(args[1])
		if err != nil {
			requests = 1
		}
		port, err = strconv.Atoi(args[2])

		if err != nil {
			port = common.DEF_SERVER_PORT
		}
	}

	client.ExecuteClient(requests, port)

}
