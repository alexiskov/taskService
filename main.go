package main

import (
	"fmt"
	"os"
	"tasvs/httpserver"
	"tasvs/sqlp"
)

var (
	HTTPtoDb  = make(chan httpserver.ToDB)
	HTTPtoWeb = make(chan httpserver.FromDB)
	SQLtoWeb  = make(chan sqlp.FromDB)
	SQLtoDB   = make(chan sqlp.ToDB)
)

func main() {
	fmt.Println(HTTPtoDb)
	go mutator(&HTTPtoDb, &HTTPtoWeb, &SQLtoDB, &SQLtoWeb)

	go func() {
		err := sqlp.New("", &SQLtoDB, &SQLtoWeb)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	}()

	err := httpserver.Start(":8080", &HTTPtoWeb, &HTTPtoDb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server initialization error!\n %+v", err)
	}

}

func mutator(fromHTTP *chan httpserver.ToDB, toHTTP *chan httpserver.FromDB, toDB *chan sqlp.ToDB, fromDB *chan sqlp.FromDB) error {
	for {
		select {
		case todb := <-*fromHTTP:
			fmt.Println(todb)
		}
	}
}
