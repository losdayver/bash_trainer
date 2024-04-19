package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/losdayver/bash_trainer/handlers"
	"github.com/losdayver/bash_trainer/persistence"
)

// TO BE REPLACED WITH CORRECT ONES
var correctCreds handlers.Creds = handlers.Creds{
	Username: "",
	Password: "",
}

// Initialising tests
func TestInit(t *testing.T) {
	os.Chdir("./../")
	persistence.Init()
}

// Tests login ability
func TestPostLoginHandler(t *testing.T) {
	testcases := []struct {
		creds          handlers.Creds
		expectedStatus int
	}{
		{correctCreds, 200}, // Replace with real credentials
		{handlers.Creds{Username: "nobody", Password: "wrong_password"}, 400},
		{handlers.Creds{Username: "", Password: ""}, 400},
		{handlers.Creds{Username: " ", Password: " "}, 400},
	}

	for _, tc := range testcases {
		s := httptest.NewServer(http.HandlerFunc(handlers.PostLoginHandler))

		reqBody, _ := json.Marshal(tc.creds)

		req, _ := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}

		if res.StatusCode != tc.expectedStatus {
			t.Errorf("unexpected status %d for user %s", tc.expectedStatus, tc.creds.Username)
		}

		res.Body.Close()
		s.Close()
	}
}

// Tests if server is able to send cors headers
func TestOptionsCorsHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handlers.PostLoginHandler))

	req, _ := http.NewRequest(http.MethodOptions, s.URL, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	res.Body.Close()
	s.Close()
}

// Tests if server can execute command via userToken
func TestPostCommandExecuteHandler(t *testing.T) {
	serverInstance1 := httptest.NewServer(http.HandlerFunc(handlers.PostLoginHandler))

	creds, _ := json.Marshal(correctCreds)
	reqLogin, _ := http.NewRequest(http.MethodPost, serverInstance1.URL, bytes.NewReader(creds))
	reqLogin.Header.Set("Content-Type", "application/json")
	resToken, _ := http.DefaultClient.Do(reqLogin)

	var userToken handlers.UserToken
	err := json.NewDecoder(resToken.Body).Decode(&userToken)
	if err != nil {
		t.Error(err)
	}

	resToken.Body.Close()
	serverInstance1.Close()

	serverInstance2 := httptest.NewServer(http.HandlerFunc(handlers.PostCommandExecuteHandler))

	command, _ := json.Marshal(handlers.CommandExecuteBody{Text: "ls -1", UserToken: userToken.UserToken})
	reqCommand, _ := http.NewRequest(http.MethodPost, serverInstance2.URL, bytes.NewReader(command))
	reqCommand.Header.Set("Content-Type", "application/json")
	resTask, err := http.DefaultClient.Do(reqCommand)
	if err != nil {
		t.Error(err)
	}

	var task handlers.Task

	err = json.NewDecoder(resTask.Body).Decode(&task)
	if err != nil {
		t.Error(err)
	}

	resTask.Body.Close()
	serverInstance2.Close()
}
