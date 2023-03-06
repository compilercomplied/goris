package main

import (
	"goris/common"
	"goris/server"
)

// This file is used only for debugging purposes. Entrypoints are provided in
// the `cmd` directory.

func main() {

	server.ExecuteServer(common.DEF_SERVER_PORT)

}
