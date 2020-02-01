package main

import (
	"fmt"
	"os"
	"flag"

	"golang.org/x/net/context"

	ggq "github.com/clarkezone/go-grpc-quick"
	test "github.com/clarkezone/gotest"
	"github.com/clarkezone/gotest/jamestestrpc"
	"google.golang.org/grpc"
)

var (
	createemptyconfig    bool
	
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

	ctx := ggq.SignalContext(context.Background()) //todo move into grpcquick
	helloServer := test.HelloServer{}
	foo := func(fn *grpc.Server) {
		fmt.Println("Callback register")
		jamestestrpc.RegisterJamesTestServiceServer(fn, &helloServer)
	}

	config := ggq.GetServerConfig()
	if (config == nil) {
		fmt.Println("No config found.  Use flag to create empty")
		os.Exit(1)
	}

	s := ggq.CreateServer(config)
	s.Serve(ctx, foo)
}
