package main

import (
	ggq "github.com/clarkezone/go-grpc-quick"
)

func main() {
	s := ggq.CreateServer()
	s.Serve()
}
