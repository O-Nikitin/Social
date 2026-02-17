package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

type envelopeErr struct {
	Error string `json:"error"`
}

type envelopeSuccess struct {
	Data any `json:"data"`
}

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	//TODO Set max size of incomming data to prevent attacs
	maxBytes := 1_048_578 //1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {

	return writeJSON(w, status, &envelopeErr{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {

	return writeJSON(w, status, &envelopeSuccess{Data: data})
}
