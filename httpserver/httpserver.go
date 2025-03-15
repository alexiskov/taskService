package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tasvs/sqlp"

	"github.com/gofiber/fiber/v2"
)

type ApiTask struct {
	ID          uint64
	Title       string
	Description string
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type (
	HTTPServer struct {
		HTTPport  string
		WebSocket Socket
		DbManager sqlp.DbManager
	}

	Socket struct {
		Driver     *fiber.App
		Config     *fiber.Config
		BridgePipe *PP
	}

	PP struct {
		Task      *ApiTask
		ISentToDB bool
		CRUDtype  int8
	}
)

func Start(anya any) error {
	task := &ApiTask{}

	srv := &HTTPServer{}
	srv.WebSocket.Config = &fiber.Config{}
	srv.WebSocket.Driver = fiber.New(*srv.WebSocket.Config)
	srv.WebSocket.BridgePipe.Task = task
	return nil
}

func (srv *HTTPServer) run() error {
	srv.WebSocket.Driver.Get("/tasks", srv.ShowTasks)
	srv.WebSocket.Driver.Post("/tasks/", srv.CreateTask)
	srv.WebSocket.Driver.Put("tasks/:id", srv.UpdateTask)
	srv.WebSocket.Driver.Delete("/tasks/:id", srv.DeleteTask)

	if err := srv.WebSocket.Driver.Listen(srv.HTTPport); err != nil {
		return err
	}
	return nil
}

// handler to GET
func (srv *HTTPServer) ShowTasks(fCtx *fiber.Ctx) error {
	if !srv.WebSocket.BridgePipe.ISentToDB {
		srv.WebSocket.BridgePipe.ISentToDB = true

		srv.WebSocket.BridgePipe.ISentToDB = false
	}
	fCtx.Set("Content-Type", "application/json")
	fCtx.Status(http.StatusOK).JSON(srv.WebSocket.BridgePipe.Task)

	return nil
}

// handler to POST
func (srv *HTTPServer) CreateTask(fCtx *fiber.Ctx) error {
	fCtx.Set("Content-Type", "application/json")
	task := ApiTask{}
	if err := json.Unmarshal(fCtx.Body(), &task); err != nil {
		return fmt.Errorf("json unmarshaling error: %w", err)
	}
	if err := srv.DbManager.CreateTask(toDBmutator(task)); err != nil {
		return err
	}
	return fCtx.Status(http.StatusOK).JSON(createResponse("task is created"))
}

// handler to PUT
func (srv *HTTPServer) UpdateTask(fCtx *fiber.Ctx) error {
	fCtx.Set("Content-Type", "application/json")
	task := ApiTask{}
	if err := json.Unmarshal(fCtx.Body(), &task); err != nil {
		return fCtx.Status(http.StatusUnprocessableEntity).JSON(createResponse("json parsing error"))
	}

	return fCtx.Status(http.StatusOK).JSON(createResponse("task is updated"))
}

// handler to DELETE
func (srv *HTTPServer) DeleteTask(fCtx *fiber.Ctx) error {
	return nil
}

// fiber.Map builder
func createResponse(a any) *fiber.Map {
	return &fiber.Map{"message": a}
}

func toResponsemutator(t sqlp.Task) ApiTask {
	return ApiTask{ID: t.ID, Title: t.Title, Description: t.Desc, CreatedAt: t.Created.Format("2006-01-02 15:04:05"), UpdatedAt: t.Updated.Format("2006-01-02 15:04:05")}
}

func toDBmutator(t ApiTask) sqlp.Task {
	return sqlp.Task{ID: t.ID, Title: t.Title, Desc: t.Description}
}
