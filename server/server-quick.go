package main

import (
	"fmt"

	ggq "github.com/clarkezone/go-grpc-quick"
	test "github.com/clarkezone/gotest"
	"github.com/clarkezone/gotest/jamestestrpc"
	"google.golang.org/grpc"
)

func main() {
	helloServer := test.HelloServer{}
	foo := func(fn *grpc.Server) {
		fmt.Println("Callback register")
		jamestestrpc.RegisterJamesTestServiceServer(fn, &helloServer)
	}

	s := ggq.CreateServer()
	s.Serve(foo)
}
