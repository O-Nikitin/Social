package main

import (
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

func (app *application) internalServerError(
	w http.ResponseWriter, r *http.Request, err error) {

	_, file, line, _ := runtime.Caller(1)
	log.Printf("[%s:%d] internal server error: %s path: %s err: %s",
		filepath.Base(file), line, r.Method, r.URL.Path, err)
	// log.Printf(
	// 	"internal server error: %s path: %s err: %s",
	// 	r.Method, r.URL.Path, err.Error())
	writeJSONError(
		w,
		http.StatusInternalServerError,
		"the server encountered a problem")
}

func (app *application) badRequestResponse(
	w http.ResponseWriter, r *http.Request, err error) {

	_, file, line, _ := runtime.Caller(1)
	log.Printf("[%s:%d] bad request error: %s path: %s err: %s",
		filepath.Base(file), line, r.Method, r.URL.Path, err)

	writeJSONError(
		w,
		http.StatusBadRequest,
		err.Error())
}

func (app *application) conflictResponse(
	w http.ResponseWriter, r *http.Request, err error) {

	_, file, line, _ := runtime.Caller(1)
	log.Printf("[%s:%d] conflict error: %s path: %s err: %s",
		filepath.Base(file), line, r.Method, r.URL.Path, err)

	writeJSONError(
		w,
		http.StatusConflict,
		err.Error())
}

func (app *application) notFoundResponse(
	w http.ResponseWriter, r *http.Request, err error) {

	_, file, line, _ := runtime.Caller(1)
	log.Printf("[%s:%d] not found error: %s path: %s err: %s",
		filepath.Base(file), line, r.Method, r.URL.Path, err)

	writeJSONError(
		w,
		http.StatusNotFound,
		"not found")
}
