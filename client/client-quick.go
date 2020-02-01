package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	ggq "github.com/clarkezone/go-grpc-quick"
	jamestestrpc "github.com/clarkezone/gotest/jamestestrpc"
)

var (
	createemptyconfig bool
)

func init() {
	flag.BoolVar(&createemptyconfig, "c", false, "create an empty configurtion file")
	flag.Parse()
}

func main() {
	if createemptyconfig {
		ggq.CreateEmptyServerConfig()
		os.Exit(0)
	}

	config := ggq.GetClientConfig()
	if config == nil {
		fmt.Println("No config found.  Use flag to create empty")
		os.Exit(1)
	}

	client := ggq.CreateClient(config)
	client.Connect()

	foo := jamestestrpc.NewJamesTestServiceClient(client.Connection)

	result, err := foo.SayHello(context.Background(), &jamestestrpc.TheHello{})

	if err != nil {
		log.Fatalf("Error calling RPC: %v", err)
	}

	if result == nil {
		log.Fatal("No error but Result was nil")
	}

	fmt.Println(result.Jamesmessage)

	defer client.Disconnect()
}
