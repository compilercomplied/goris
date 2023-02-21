package main

import (
	"fmt"
	"goris/client"
	"goris/server"
	"runtime"
)

func main() {
	fmt.Println("run server")
	go server.ExecuteServer()
	fmt.Println("run client")
	go client.ExecuteClient()
	runtime.Goexit()

}
