package sqlp

import (
	"context"
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
		ID      uint64    `json:"id"`
		Title   string    `json:"title"`
		Desc    string    `json:"description"`
		Status  string    `json:"status"`
		Created time.Time `json:"created_at"`
		Updated time.Time `json:"updated_at"`
	}

	FromDB struct {
		Task      *[]Task
		NumThread int64
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
	db.run(inc, out)
	return
}

func (db *dB) run(inc *chan ToDB, out *chan FromDB) error {
	for {
		select {
		case incomingData := <-*inc:
			switch incomingData.CRUDtype {
			case 1:
				db.ShowTasks()
			}
		}
	}
}

// method SELECT
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

// method CREATE
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

func (db *dB) UpdateTask(task Task) error {
	//	date:
	db.Socket.QueryRow(context.Background(), "update (`title`, `description`, `status`, `updated_at`) IN `tableName` where `id`=$1", task.ID, task.Title, task.Desc, task.Status, time.Now())
	return nil
}

//func (db *dB)
