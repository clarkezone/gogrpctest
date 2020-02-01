package main

import (
	test "github.com/clarkezone/gotest"
)

func mainr() {
	bs := test.CreateBackend()
	bs.StartclientStreaming()
	//bs.StartclientSecure()
	//bs.StartclientStreamingSecure()
}
