package main

import (
	"fmt"

	test "github.com/clarkezone/gotest"
)

func main() {
	var c test.Conf
	c.GetConf()
	fmt.Println("Hello")

	test.Startclientsecure(c.TlsServerName, c.ClientPort)
}
