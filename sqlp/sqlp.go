package sqlp

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbManager interface {
	ShowTasks() (Tasks, error)
	UpdateTask(newTask Task) (task *Task, err error)
	CreateTask(task Task) error
}

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

	Tasks []Task
)

// storage manager initialization
// return interface or error
func New(socket string) (db DbManager, err error) {
	database := dB{}
	conf, err := pgxpool.ParseConfig(socket)
	if err != nil {
		return nil, fmt.Errorf("database credentials parsing error: %w", err)
	}
	database.Socket, err = pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("database connection pool making error: %w\n", err)
	}
	db = &database
	return
}

func Run(mgr *DbManager, anya any) {

}

// method SELECT
func (db *dB) ShowTasks() (Tasks, error) {
	tx, err := db.Socket.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database socket initialization error:\n %w", err)
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), "select * from `tableName`")
	if err != nil {
		return nil, fmt.Errorf("error of task list selection:\n %w", err)
	}

	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return nil, fmt.Errorf("parse rows to list of task error: %w", err)
	}
	return tasks, nil
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

func (db *dB) UpdateTask(task Task) (*Task, error) {
	//	date:
	db.Socket.QueryRow(context.Background(), "update (`title`, `description`, `status`, `updated_at`) IN `tableName` where `id`=$1", task.ID, task.Title, task.Desc, task.Status, time.Now())
	return nil, nil
}

//func (db *dB)
