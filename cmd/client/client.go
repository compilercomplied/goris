package main

import (
	"goris/client"
	"goris/protocol"
	"os"
)

func parseArgs(args []string) (*protocol.ProtocolRequest, error) {

	// Recklessly ignore fail scenarios. This stunt is performed by experts,
	// do not emulate at home.
	if len(args) == 2 {
		return protocol.NewProtocolRequest(args[0], args[1], nil)
	} else {
		val := args[2]
		return protocol.NewProtocolRequest(args[0], args[1], &val)
	}

}

func main() {

	req, err := parseArgs(os.Args[1:])

	if err != nil {
		panic(err)
	}

	client.ExecuteClient(req)

}
