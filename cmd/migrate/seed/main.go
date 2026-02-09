package main

import (
	"log"

	"github.com/O-Nikitin/Social/internal/db"
	"github.com/O-Nikitin/Social/internal/store"
)

// This is a help func to create some test data and write it into DB
func main() {
	conn, err := db.New(
		"postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable",
		3,
		3,
		"15m")
	if err != nil {
		log.Panic(err.Error())
	} else {
		log.Println("DB connected!")
	}
	defer conn.Close()
	store := store.NewStorage(conn)

	db.Seed(store, conn)
}
