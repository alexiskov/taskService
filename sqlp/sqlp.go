package sqlp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dB struct {
	Socket *pgxpool.Pool
}

type (
	Task struct {
		ID      uint64
		Title   string
		Desc    string
		Status  string
		Created time.Time
		Updated time.Time
	}

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

func New(socket string, inc *chan ToDB, out *chan FromDB) (err error) {
	db := dB{}
	conf, err := pgxpool.ParseConfig(socket)
	if err != nil {
		return fmt.Errorf("database credentials parsing error: %w", err)
	}
	db.Socket, err = pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return fmt.Errorf("database connection pool making error: %w\n", err)
	}
	if err = db.run(inc, out); err != nil {
		return
	}
	return
}

// получает данные из канала
// выполняет операции CRUD над базой данных
// номер операции соответствует порядковому номеру символа аббривиатуры CRUD, начиная с 1...
func (db *dB) run(inc *chan ToDB, out *chan FromDB) error {
	for {
		select {
		case incomingData := <-*inc:
			switch incomingData.CRUDtype {
			case 1:
				task := Task{Title: incomingData.Task.Title, Desc: incomingData.Task.Desc}
				if err := db.CreateTask(task); err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//logger
					continue
				}
				*out <- FromDB{NumThread: incomingData.NumThread, IsOk: true}
			case 2:
				tasks, err := db.ShowTasks()
				if err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread}
					//loger
					continue
				}
				*out <- FromDB{Task: &tasks, NumThread: incomingData.NumThread, IsOk: true}
			case 3:
				task := Task{ID: incomingData.Task.ID, Title: incomingData.Task.Title, Desc: incomingData.Task.Desc, Status: incomingData.Task.Status, Updated: time.Now()}
				resp, err := db.UpdateTask(task)
				if err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread} //logger
					continue
				}
				tasks := []Task{}
				tasks = append(tasks, resp)
				*out <- FromDB{Task: &tasks, NumThread: incomingData.NumThread, IsOk: true}
			case 4:
				task := Task{Title: incomingData.Task.Title, Desc: incomingData.Task.Desc}
				if err := db.RemoveTask(task); err != nil {
					var empty []Task
					*out <- FromDB{Task: &empty, NumThread: incomingData.NumThread} //logger
					continue
				}
				*out <- FromDB{NumThread: incomingData.NumThread, IsOk: true}
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

	rows, err := tx.Query(context.Background(), "select * from `tableName`")
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

	var id uint64
	if err = tx.QueryRow(context.Background(), "INSERT INTO tasks (title,description,) VALUES($1,$2)", task.Title, task.Desc).Scan(&id); err != nil {
		return fmt.Errorf("task creating error: %w", err)
	}
	return nil
}

// ...UPDATE
func (db *dB) UpdateTask(task Task) (Task, error) {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return Task{}, fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())
	taskResponse := Task{}
	comTag, err := tx.Exec(context.Background(), "update (`title`, `description`, `status`, `updated_at`) IN `tableName` where `id`=$1", task.ID, task.Title, task.Desc, task.Status, task.Updated.Unix())
	if err != nil {
		return Task{}, err
	}
	if comTag.RowsAffected() == 0 {
		return Task{}, fmt.Errorf("DBdata updating error: %w", err)
	}
	return taskResponse, nil
}

// ...DELETE
func (db *dB) RemoveTask(task Task) error {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	comTag, err := tx.Exec(context.Background(), "DELETE FROM tasks WHERE id=$1", task.ID)
	if err != nil {
		return fmt.Errorf("DBdata deleting error: %w", err)
	}
	if comTag.RowsAffected() == 0 {
		return errors.New("DBdata deleting error: Record is Not Deleted")
	}
	return nil
}
