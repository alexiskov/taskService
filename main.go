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
	go mutator(&HTTPtoDb, &HTTPtoWeb, &SQLtoDB, &SQLtoWeb)

	go func() error {
		err := sqlp.New("postgres://pguser:secret!@localhost:5432/db_1", &SQLtoDB, &SQLtoWeb)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return err
		}
		return nil
	}()

	err := httpserver.Start(":8080", &HTTPtoWeb, &HTTPtoDb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server initialization error!\n %+v", err)
	}

}

func mutator(fromHTTP *chan httpserver.ToDB, toHTTP *chan httpserver.FromDB, toDBchan *chan sqlp.ToDB, fromDB *chan sqlp.FromDB) {
	go func() {
		for {
			select {
			case todb := <-*fromHTTP:
				*toDBchan <- sqlp.ToDB{
					Task:      &sqlp.Task{ID: todb.Task.ID, Title: todb.Task.Title, Desc: todb.Task.Description},
					NumThread: todb.NumThread,
					CRUDtype:  todb.CRUDtype,
				}
			}
		}
	}()

	for {
		select {
		case fromDbData := <-*fromDB:
			var tasks []httpserver.ApiTask
			for _, t := range *fromDbData.Task {
				task := httpserver.ApiTask{ID: t.ID, Title: t.Title, Description: t.Desc, Status: t.Status, CreatedAt: t.Created.Format("2006-01-02 15:04:05"), UpdatedAt: t.Updated.Format("2006-01-02 15:04:05")}
				tasks = append(tasks, task)
			}
			*toHTTP <- httpserver.FromDB{
				Task:      tasks,
				NumThread: fromDbData.NumThread,
				IsOk:      fromDbData.IsOk,
			}
		}
	}
}
