package main

import (
	"fmt"
	"log"
	"os"
	"tasvs/confreader"
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
	//mainConfig_loading
	config, err := confreader.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	go mutatorHTTPandSQL(&HTTPtoDb, &HTTPtoWeb, &SQLtoDB, &SQLtoWeb)

	go func() error {
		err := sqlp.New(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.DB.Username, config.DB.Password, config.DB.Ip, config.DB.Port, config.DB.DBname), &SQLtoDB, &SQLtoWeb)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return err
		}
		return nil
	}()

	err = httpserver.Start(config.WebSRV.Port, &HTTPtoWeb, &HTTPtoDb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server initialization error!\n %+v", err)
	}

}

//DATAconverter
func mutatorHTTPandSQL(fromHTTP *chan httpserver.ToDB, toHTTP *chan httpserver.FromDB, toDBchan *chan sqlp.ToDB, fromDB *chan sqlp.FromDB) {
	// HTTPtoSQL_thread
	go func() {
		for {
			select {
			case todb := <-*fromHTTP:
				*toDBchan <- sqlp.ToDB{
					Task:      &sqlp.Task{ID: todb.Task.ID, Title: todb.Task.Title, Desc: todb.Task.Description, Status: todb.Task.Status},
					NumThread: todb.NumThread,
					CRUDtype:  todb.CRUDtype,
				}
			}
		}
	}()

	// SQLtoHTTP_thread
	for {
		select {
		case fromDbData := <-*fromDB:
			var tasks []httpserver.ApiTask
			for _, t := range *fromDbData.Task {
				task := httpserver.ApiTask{ID: t.ID, Title: t.Title, Description: t.Desc, Status: t.Status, CreatedAt: t.Created.Local().Format("2006-01-02 15:04:05"), UpdatedAt: t.Updated.Local().Format("2006-01-02 15:04:05")}
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
