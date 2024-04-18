package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/losdayver/bash_trainer/persistence"
)

func OptionsCorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Origin", persistence.Config.Origin)

	w.WriteHeader(204)
}

var Tasks TaskHashMap = TaskHashMap{
	Hashmap: make(map[uuid.UUID]*Task),
	Mutex:   &sync.Mutex{},
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

var UserSessions UserSessionHashMap = UserSessionHashMap{
	Hashmap: make(map[string]uuid.UUID),
	Mutex:   &sync.Mutex{},
}

func (userSessions UserSessionHashMap) AddOrRenew(username string) uuid.UUID {
	defer userSessions.Mutex.Unlock()

	userToken, _ := uuid.NewV4()
	userSessions.Mutex.Lock()
	userSessions.Hashmap[username] = userToken
	return userToken
}

func (userSessions UserSessionHashMap) TestExists(userTokenStr string) bool {
	userToken, err := uuid.FromString(userTokenStr)

	if err != nil {
		return false
	}

	for _, value := range userSessions.Hashmap {
		if value == userToken {
			return true
		}
	}

	return false
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	var body Creds

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authenticated, err := persistence.Authenticate(body.Username, body.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if authenticated {
		userToken := UserSessions.AddOrRenew(body.Username)

		userTokenBody := UserToken{UserToken: userToken.String()}
		jsonData, err := json.Marshal(userTokenBody)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	} else {
		http.Error(w, "Invalid Credentials", http.StatusBadRequest)
	}
}

func PostCommandExecuteHandler(w http.ResponseWriter, r *http.Request) {
	var body CommandExecuteBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !UserSessions.TestExists(body.UserToken) {
		http.Error(w, "Invalid Token", http.StatusBadRequest)
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
		time.Sleep(time.Second * 1)

		cmd := exec.Command("/bin/bash", "-c", "cd ./public/library && "+body.Text)

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
}

func GetCommandPalette(w http.ResponseWriter, r *http.Request) {
	var body GetCommandPacket

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !UserSessions.TestExists(body.UserToken) {
		http.Error(w, "Invalid Token", http.StatusBadRequest)
		return
	}

	commandList, err := persistence.QueryCommands(body.Username)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	commandPalette := CommandPalettePacket{commandList}

	jsonData, err := json.Marshal(commandPalette)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(jsonData)
}

func PostCommandSaveHandler(w http.ResponseWriter, r *http.Request) {
	var body SaveCommandPacket

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !UserSessions.TestExists(body.UserToken) {
		http.Error(w, "Invalid Token", http.StatusBadRequest)
		return
	}

	err1 := persistence.SaveCommand(body.Username, body.Command)

	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(200)
}
