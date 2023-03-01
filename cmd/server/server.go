package main

import (
	"goris/common"
	"goris/server"
	"os"
	"strconv"
)

func main() {

	args := os.Args

	var port int
	var err error = nil

	// First arg is a reference to the binary itself.
	if len(args) == 1 {
		port = common.DEF_SERVER_PORT
	} else {
		port, err = strconv.Atoi(args[1])
		if err != nil {
			port = common.DEF_SERVER_PORT
		}
	}

	server.ExecuteServer(port)
}
