package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	TASK_RUNNING = 0
	TASK_DONE    = 1
	TASK_FAILED  = 2
)

type Task struct {
	TaskToken uuid.UUID
	Status    int
	Output    string
}

var Tasks TaskHashMap = TaskHashMap{
	Hashmap: make(map[uuid.UUID]*Task),
	Mutex:   &sync.Mutex{},
}

type TaskHashMap struct {
	Hashmap map[uuid.UUID]*Task
	Mutex   *sync.Mutex
}

func (taskHashMap TaskHashMap) AppendTask(token uuid.UUID, task *Task) {
	taskHashMap.Mutex.Lock()
	taskHashMap.Hashmap[token] = task
	defer taskHashMap.Mutex.Unlock()
}

func (taskHashMap TaskHashMap) RemoveTask(token uuid.UUID) error {
	defer taskHashMap.Mutex.Unlock()

	taskHashMap.Mutex.Lock()

	if _, ok := taskHashMap.Hashmap[token]; ok {
		delete(taskHashMap.Hashmap, token)
		return nil
	} else {
		return errors.New("invalid Token")
	}
}

func (taskHashMap TaskHashMap) GetTask(token uuid.UUID) *Task {
	defer taskHashMap.Mutex.Unlock()
	taskHashMap.Mutex.Lock()
	return taskHashMap.Hashmap[token]
}

func (taskHashMap TaskHashMap) ChangeStatus(token uuid.UUID, status int, output string) {
	defer taskHashMap.Mutex.Unlock()
	taskHashMap.Mutex.Lock()
	taskHashMap.Hashmap[token].Status = status
	taskHashMap.Hashmap[token].Output = output
}

var Origin string = "localhost:4000"

type CommandExecuteBody struct {
	Text      string
	UserToken string
}

type CommandPaletteBody struct {
	Success  bool
	Commands []string
}

func ApiWrapper(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Origin", Origin)

		handler(w, r)
	}
}

func OptionsCorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Origin", Origin)

	w.WriteHeader(204)
}

func PostCommandExecuteHandler(w http.ResponseWriter, r *http.Request) {
	var body CommandExecuteBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := uuid.NewV4()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := &Task{
		TaskToken: token,
		Status:    TASK_RUNNING,
		Output:    "",
	}

	Tasks.AppendTask(token, task)

	jsonData, err := json.Marshal(task)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(jsonData)

	go func() {
		time.Sleep(time.Second * 10)

		cmd := exec.Command("bash", "-c", body.Text)

		stdout, err := cmd.Output()

		if err != nil {
			Tasks.ChangeStatus(token, TASK_FAILED, err.Error()+"\n"+string(stdout))
			return
		}

		Tasks.ChangeStatus(token, TASK_DONE, string(stdout))
	}()
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	tokenParam := r.PathValue("token")

	if tokenParam == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	token, err := uuid.FromString(tokenParam)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := Tasks.GetTask(token)

	if task == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(task)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(jsonData)

	// Tasks.RemoveTask(token)
}

func GetCommandPalette(w http.ResponseWriter, r *http.Request) {

}

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public/"))

	// Map the URL path "/home/" to "index.html"
	mux.Handle("/public/", http.StripPrefix("/public/", fs))

	mux.HandleFunc("POST /api/command/execute/{$}", ApiWrapper(PostCommandExecuteHandler))
	mux.HandleFunc("GET /api/task/{token}", ApiWrapper(GetTaskHandler))
	mux.HandleFunc("OPTIONS /api/", OptionsCorsHandler)

	// Serving index.html
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/views/index.html")
	})

	log.Fatal(http.ListenAndServe(Origin, mux))
}
