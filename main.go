package main

import (
	"fmt"
	"goris/client"
	"goris/server"
	"runtime"
	"time"
)

func main() {
	fmt.Println("run server")
	go server.ExecuteServer()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("run client")
	go client.ExecuteClient()
	// This is breaking due to how goroutines are designed.
	// It'll get fixed later
	// TODO: goroutines
	runtime.Goexit()

}
