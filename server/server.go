package main

import (
	test "github.com/clarkezone/gotest"
)

func main() {
	var c test.Conf
	c.GetConf()

	test.ServegRPC(c.TlsServerName, c.ServerPort)
}
