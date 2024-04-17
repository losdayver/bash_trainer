package handlers

import (
	"sync"

	"github.com/gofrs/uuid/v5"
)

type UserSessionHashMap struct {
	Hashmap map[string]uuid.UUID
	Mutex   *sync.Mutex
}

/*
	RECEIVED TYPES
*/

// JSON is received when user tries to log in
type Creds struct {
	Username string
	Password string
}

/*
	SENT TYPES
*/

// JSON is sent on successful auth
type UserToken struct {
	UserToken string
}
