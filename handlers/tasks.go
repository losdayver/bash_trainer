package handlers

import (
	"sync"

	"github.com/gofrs/uuid/v5"
)

const (
	TASK_RUNNING = 0
	TASK_DONE    = 1
	TASK_FAILED  = 2
)

// Hashmap for storing running/done/failed tasks
type TaskHashMap struct {
	Hashmap map[uuid.UUID]*Task
	Mutex   *sync.Mutex
}

/*
	SENT TYPES
*/

// JSON represents task in a hashmap
type Task struct {
	TaskToken uuid.UUID
	Status    int
	Output    string
}

// JSON is sent to populate user's command palette
type CommandPaletteBody struct {
	Success  bool
	Commands []string
}

/*
	RECEIVED TYPES
*/

// JSON is received when user presses "Execute"
type CommandExecuteBody struct {
	Text      string
	UserToken string
}
