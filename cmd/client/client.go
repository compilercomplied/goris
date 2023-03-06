package main

import (
	"goris/client"
	"goris/common"
	"os"
)

func parseArgs(args []string) (*common.ProtocolRequest, error) {

	// Recklessly ignore fail scenarios. This stunt is performed by experts, 
	// do not emulate at home.
	if len(args) == 2 {
		return common.NewProtocolRequest(args[0], args[1], nil)
	} else {
		val := args[2]
		return common.NewProtocolRequest(args[0], args[1], &val)
	}

}

func main() {

	req,err := parseArgs(os.Args[1:])

	if err != nil { panic(err)}



	client.ExecuteClient(req)

}

