package sqlp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//DB_manager
type dB struct {
	Socket *pgxpool.Pool
}

type (
	//DATA_template
	Task struct {
		ID      uint64    `db:"id, omitempty"`
		Title   string    `db:"title, omitempty"`
		Desc    string    `db:"description, omitempty"`
		Status  string    `db:"status, omitempty"`
		Created time.Time `db:"created_at, omitempty"`
		Updated time.Time `db:"updated_at, omitempty"`
	}

	//incoming && outgoing CHANNEL_structs
	FromDB struct {
		Task      *[]Task
		NumThread int64
		IsOk      bool
	}

	ToDB struct {
		Task      *Task
		CRUDtype  int8
		NumThread int64
	}
)

//SQL_worker 
//use incomming and outgoing channels in arg
func New(dbConfig string, inc *chan ToDB, out *chan FromDB) (err error) {
	//LOAD config to external SQL package driver
	db := dB{}
	conf, err := pgxpool.ParseConfig(dbConfig)
	conf.ConnConfig.TLSConfig = nil
	if err != nil {
		return fmt.Errorf("database credentials parsing error: %w", err)
	}
	//USE SQL external package driver of DB_manager
	db.Socket, err = pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return fmt.Errorf("database connection pool making error: %w\n", err)
	}
	//DB_manager RUNNING
	if err = db.run(inc, out); err != nil {
		return
	}
	return
}

//DB_manager
// получает данные из канала
// выполняет операции CRUD над базой данных
// номер операции соответствует порядковому номеру символа аббривиатуры CRUD, начиная с 1...
func (db *dB) run(inc *chan ToDB, out *chan FromDB) error {
	// channel_thread LOOP start
	for {
		select {
		case incomingData := <-*inc:
			//CRUD_type checking
			switch incomingData.CRUDtype {
			case 1:
				task := Task{Title: incomingData.Task.Title, Desc: incomingData.Task.Desc}
				//try task writing
				if err := db.CreateTask(task); err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//logger from db
					continue
				}
				var tasks []Task
				*out <- FromDB{Task: &tasks, NumThread: incomingData.NumThread, IsOk: true}
			case 2:
				tasks, err := db.ShowTasks()
				//try taskList getting
				if err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//logger from db
					continue
				}
				*out <- FromDB{Task: &tasks, NumThread: incomingData.NumThread, IsOk: true}
			case 3:
				task := Task{ID: incomingData.Task.ID, Title: incomingData.Task.Title, Desc: incomingData.Task.Desc, Status: incomingData.Task.Status, Updated: time.Now().UTC()}
				//try task updating
				err := db.UpdateTask(task)
				var empty []Task
				if err != nil {
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//logger from db
					continue
				}
				*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread, IsOk: true}
			case 4:
				task := Task{ID: incomingData.Task.ID, Title: incomingData.Task.Title, Desc: incomingData.Task.Desc}
				var empty []Task
				if err := db.RemoveTask(task); err != nil {
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//logger from db
					//-->example
					log.Println(err)
					//<--example
					continue
				}
				*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread, IsOk: true}
			}
		}
	}
}

// method SQL SELECT
func (db *dB) ShowTasks() (tasks []Task, err error) {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), "select * from tasks")
	if err != nil {
		return nil, fmt.Errorf("error of task list selection:\n %w", err)
	}

	tasks, err = pgx.CollectRows(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return nil, fmt.Errorf("parse rows to list of task error: %w", err)
	}
	return
}

// ...CREATE
func (db *dB) CreateTask(task Task) error {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	query := "INSERT INTO tasks(title, description) VALUES(@title,@description)"
	args := pgx.NamedArgs{
		"title":       task.Title,
		"description": task.Desc,
	}
	if _, err := tx.Exec(context.Background(), query, args); err != nil {
		return fmt.Errorf("task creating error: %w", err)
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("task creating error: %w", err)
	}
	return nil
}

// ...UPDATE
func (db *dB) UpdateTask(task Task) error {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	query := "UPDATE tasks SET title=@title, description=@description, status=@status, updated_at=@updated_at where id=@id"
	args := pgx.NamedArgs{
		"id":          task.ID,
		"title":       task.Title,
		"description": task.Desc,
		"status":      task.Status,
		"updated_at":  task.Updated,
	}
	_, err = tx.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("task updating error: %w", err)
	}
	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("task updating error: %w", err)
	}
	return nil
}

// ...DELETE
func (db *dB) RemoveTask(task Task) error {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	query := "DELETE FROM tasks WHERE id=@id"
	args := pgx.NamedArgs{"id": task.ID}

	_, err = tx.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("DBdata deleting error: %w", err)
	}
	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("DBdata deleting error: %w", err)
	}
	return nil
}
