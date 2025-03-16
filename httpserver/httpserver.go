package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

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
		HTTPport           string
		WebSocket          Socket
		dbResponseReciever map[int64]FromDB
	}

	Socket struct {
		Driver     *fiber.App
		Config     *fiber.Config
		BridgePipe *PP
	}

	PP struct {
		Out      *chan ToDB
		Incoming *chan FromDB
	}

	ToDB struct {
		Task      ApiTask
		CRUDtype  int8
		NumThread int64
	}

	FromDB struct {
		Task      ApiTask
		NumThread int64
	}
)

func Start(inc *chan FromDB, out *chan ToDB) error {
	srv := &HTTPServer{}
	srv.WebSocket.Config = &fiber.Config{}
	srv.WebSocket.Driver = fiber.New(*srv.WebSocket.Config)
	srv.WebSocket.BridgePipe = &PP{out, inc}
	return nil
}

func (srv *HTTPServer) Run() error {
	go srv.recieveMGR(srv.WebSocket.BridgePipe.Incoming)
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

	fCtx.Set("Content-Type", "application/json")
	//
	//over chan to db data sent
	//
	fCtx.Status(http.StatusOK).JSON()

	return nil
}

// handler to POST
func (srv *HTTPServer) CreateTask(fCtx *fiber.Ctx) error {
	fCtx.Set("Content-Type", "application/json")
	task := ApiTask{}
	if err := json.Unmarshal(fCtx.Body(), &task); err != nil {
		return fmt.Errorf("json unmarshaling error: %w", err)
	}
	//to db
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

func (src *HTTPServer) recieveMGR(inc *chan FromDB) {
	recieverOfResponseFromDB := make(map[int64]FromDB, 1000)
	var mutex sync.Mutex
	for {
		select {
		case fromDBdata := <-*inc:
			mutex.Lock()
			recieverOfResponseFromDB[fromDBdata.NumThread] = fromDBdata
			mutex.Unlock()
		}
	}
}
