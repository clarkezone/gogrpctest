package main

import (
	test "github.com/clarkezone/gotest"
)

func main() {
	//var c test.Conf
	//c.GetConf()
	//test.StartclientStreaming(c.TlsServerName, c.ClientPort)

	bs := test.CreateBackend()
	bs.StartclientStreaming()
}
