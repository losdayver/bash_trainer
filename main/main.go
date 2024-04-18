package main

import (
	"log"
	"net/http"

	"github.com/losdayver/bash_trainer/handlers"
	"github.com/losdayver/bash_trainer/persistence"
)

func ApiWrapper(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Origin", persistence.Config.Origin)

		handler(w, r)
	}
}

func main() {
	persistence.Init()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public/"))

	mux.Handle("/public/", http.StripPrefix("/public/", fs))

	mux.HandleFunc("POST /api/command/execute/{$}", ApiWrapper(handlers.PostCommandExecuteHandler))
	mux.HandleFunc("POST /api/command/save/{$}", ApiWrapper(handlers.PostCommandSaveHandler))
	mux.HandleFunc("POST /api/palette/{$}", ApiWrapper(handlers.GetCommandPalette))
	mux.HandleFunc("GET /api/task/{token}", ApiWrapper(handlers.GetTaskHandler))
	mux.HandleFunc("POST /api/login/{$}", ApiWrapper(handlers.PostLoginHandler))
	mux.HandleFunc("OPTIONS /api/", handlers.OptionsCorsHandler)

	// Serving index.html
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/static/views/index.html")
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:4000", mux))
}
