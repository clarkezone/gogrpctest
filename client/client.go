package main

import (
	test "github.com/clarkezone/gotest"
)

func main() {
	bs := test.CreateBackend()
	//bs.StartclientStreaming()
	//bs.StartclientSecure()
	bs.StartclientStreamingSecure()
}
