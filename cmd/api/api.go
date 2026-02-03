package main

import (
	"log"
	"net/http"
	"time"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	serverWriteTimeout time.Duration = 30
	serverReadTimeout  time.Duration = 20
)

type config struct {
	addr string
	db   dbConfig
}
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxidleTime  string
}

type application struct {
	config config
	store  store.Storage
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//Adds information about each received request
	r.Use(middleware.Logger)
	//Recovers from panic and return internal server error
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(1 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.Post("/health", app.healthCheckHandler)
	})

	//here we are going to have a lot of endpoints
	//posts
	//users
	//etc
	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * serverWriteTimeout,
		ReadTimeout:  time.Second * serverReadTimeout,
		IdleTimeout:  time.Minute,
	}
	log.Printf("Server has started at %s ", app.config.addr)
	return srv.ListenAndServe()
}
