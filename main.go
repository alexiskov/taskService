package main

import (
	"fmt"
	"os"
	"tasvs/bridge"
	"tasvs/httpserver"
	"tasvs/sqlp"
)

type (
	ToDbMemCash struct {
	}
	ToHttpMemCash struct {
	}
)

func main() {
	toDb := &ToDbMemCash{}
	toHTTP := &ToHttpMemCash{}

	bridge.Run(toHTTP, toDb)

	err := httpserver.Start(toHTTP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server initialization error!\n %+v", err)
	}

	dnManager, err := sqlp.New("")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}
}
