package httpserver

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
)

var TASK_STATS = [3]string{"new", "in_progress", "done"}

type ApiTask struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	IsOk        bool   `json:"-"`
}

type (
	HTTPServer struct {
		HTTPport           string
		WebSocket          Socket
		dbResponseReciever map[int64]FromDB
		threadPool         map[int64]bool
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
		Task      []ApiTask
		NumThread int64
		IsOk      bool
	}
)

func Start(port string, inc *chan FromDB, out *chan ToDB) error {
	srv := &HTTPServer{HTTPport: port}
	srv.WebSocket.Config = &fiber.Config{}
	srv.WebSocket.Driver = fiber.New(*srv.WebSocket.Config)
	srv.WebSocket.BridgePipe = &PP{out, inc}
	srv.threadPool = make(map[int64]bool, 1000)
	if err := srv.run(); err != nil {
		return err
	}
	return nil
}

func (srv *HTTPServer) run() error {
	go srv.recieveMGR(srv.WebSocket.BridgePipe.Incoming)
	srv.WebSocket.Driver.Get("/tasks", srv.ShowTasks)
	srv.WebSocket.Driver.Post("/tasks", srv.CreateTask)
	srv.WebSocket.Driver.Put("/tasks/:id?", srv.UpdateTask)
	srv.WebSocket.Driver.Delete("/tasks/:id?", srv.DeleteTask)

	if err := srv.WebSocket.Driver.Listen(srv.HTTPport); err != nil {
		return err
	}
	return nil
}

// handler to GET
func (srv *HTTPServer) ShowTasks(fCtx *fiber.Ctx) error {
	threadData := srv.threadToDbPrepeare(ToDB{Task: ApiTask{}, CRUDtype: 2})
	//to mutator
	*srv.WebSocket.BridgePipe.Out <- threadData
	//get task from response pool by number of thread
	data := srv.dbResponseSeparator(threadData.NumThread)
	//http.response
	if !data.IsOk {
		return fCtx.Status(http.StatusUnprocessableEntity).JSON(createResponse("database processing error"))
	}
	fCtx.Set("Content-Type", "application/json")
	fCtx.Status(http.StatusOK).JSON(data.Task)
	return nil
}

// handler to POST
func (srv *HTTPServer) CreateTask(fCtx *fiber.Ctx) error {
	fCtx.Set("Content-Type", "application/json")
	task := ApiTask{}
	if err := json.Unmarshal(fCtx.Body(), &task); err != nil {
		return fmt.Errorf("json unmarshaling error: %w", err)
	}
	//to mutator
	threadData := srv.threadToDbPrepeare(ToDB{Task: task, CRUDtype: 1})
	*srv.WebSocket.BridgePipe.Out <- threadData
	//from mutator
	data := srv.dbResponseSeparator(threadData.NumThread)
	//HTTP response
	fCtx.Set("Content-Type", "application/json")
	if !data.IsOk {
		return fCtx.Status(http.StatusUnprocessableEntity).JSON(createResponse("database processing error"))
	}
	fCtx.Status(http.StatusOK).JSON(data.Task[0])
	return nil
}

// handler to PUT
func (srv *HTTPServer) UpdateTask(fCtx *fiber.Ctx) error {
	fCtx.Set("Content-Type", "application/json")
	task := ApiTask{}
	if err := json.Unmarshal(fCtx.Body(), &task); err != nil {
		return fCtx.Status(http.StatusUnprocessableEntity).JSON(createResponse("json parsing error"))
	}
	isOK := false
	for _, stauts := range TASK_STATS {
		if stauts == task.Status {
			isOK = true
		}
	}
	if !isOK {
		return fCtx.Status(http.StatusUnprocessableEntity).JSON(createResponse(fmt.Sprintf("status '%s' is not valid", task.Status)))
	}
	//to mutator
	threadData := srv.threadToDbPrepeare(ToDB{Task: task, CRUDtype: 3})
	*srv.WebSocket.BridgePipe.Out <- threadData
	//from mutator

	return fCtx.Status(http.StatusOK).JSON(createResponse("task is updated"))
}

// handler to DELETE
func (srv *HTTPServer) DeleteTask(fCtx *fiber.Ctx) error {
	//
	//
	return nil
}

// fiber.Map builder
func createResponse(a any) *fiber.Map {
	return &fiber.Map{"message": a}
}

// выбираем номер потока не всписке отправленных
// возвращает подготовленный тип для отправки по каналу в мутатор
func (srv *HTTPServer) threadToDbPrepeare(data ToDB) ToDB {
	var mutex sync.Mutex
	var numThread int64
	for {
		numThreadData := rand.Intn(1844674407370955161)
		if _, ok := srv.threadPool[int64(numThreadData)]; !ok {
			numThread = int64(numThreadData)
			mutex.Lock()
			srv.threadPool[int64(numThreadData)] = true
			mutex.Unlock()
			break
		}
	}
	data.NumThread = numThread
	return data
}

// воркер, обрабатывающий ответ от мутатора через канал
// наполняет пулл ответов от БД для веб-сервера, для последующего разбора и отправки клиенту
func (srv *HTTPServer) recieveMGR(inc *chan FromDB) {
	srv.dbResponseReciever = make(map[int64]FromDB, 1000000)
	var mutex sync.Mutex
	for {
		select {
		case fromDBdata := <-*inc:
			mutex.Lock()
			srv.dbResponseReciever[fromDBdata.NumThread] = fromDBdata
			mutex.Unlock()
		}
	}
}

// получает данные из пулла ответов от базы данных, возвращает задачи соответствующие номеру потока вызвавшего клиента
// удаляет использованные данные из накопителя
func (srv *HTTPServer) dbResponseSeparator(numThread int64) FromDB {
	var mutex sync.Mutex
	data := FromDB{}
	for {
		if val, ok := srv.dbResponseReciever[numThread]; ok {
			data = FromDB{Task: val.Task, IsOk: val.IsOk}
			mutex.Lock()
			delete(srv.dbResponseReciever, val.NumThread)
			mutex.Unlock()
			break
		}
	}
	return data
}
