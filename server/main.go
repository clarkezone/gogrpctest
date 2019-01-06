package main

import (
	test "github.com/clarkezone/gotest"
)

func main() {
	var c test.Conf
	c.GetConf()

	test.ServegRPCAutoCert(c.TlsServerName, c.ServerPort)
}
