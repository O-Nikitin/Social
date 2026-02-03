package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("I'm ok!\n"))

	//Write some data to storage
	//app.store.Posts.Create(r.Context())
}
