package main

import (
	test "github.com/clarkezone/gotest"
)

func main() {
	bs := test.CreateBackend()
	//bs.ServegRPC()
	bs.ServegRPCAutoCert()
}