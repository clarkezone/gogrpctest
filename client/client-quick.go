package main

import (
	"context"
	"fmt"
	"log"

	ggq "github.com/clarkezone/go-grpc-quick"
	jamestestrpc "github.com/clarkezone/gotest/jamestestrpc"
)

func main() {
	client := ggq.CreateClient()
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
