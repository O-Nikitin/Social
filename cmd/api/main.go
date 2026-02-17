package main

import (
	"log"

	"github.com/O-Nikitin/Social/internal/db"
	"github.com/O-Nikitin/Social/internal/env"
	"github.com/O-Nikitin/Social/internal/store"
)

const version = "0.0.1"

//	@title			Social API
//	@description	API for Social app.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1
//
//@securitydefinitions.apikey ApiKeyAuth
//@in			header
//@name 		Authorization
//@description

func main() {
	//flag is used to print file name and line in log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr: env.GetString(
				"DB_ADDR",
				"postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxidleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:    env.GetString("ENV", "development"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:3000")}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxidleTime)
	if err != nil {
		log.Panic(err.Error())
	} else {
		log.Println("DB connected!")
	}
	defer db.Close()
	store := store.NewStorage(db)

	app := application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
