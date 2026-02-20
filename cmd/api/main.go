package main

import (
	"time"

	"github.com/O-Nikitin/Social/internal/db"
	"github.com/O-Nikitin/Social/internal/env"
	"github.com/O-Nikitin/Social/internal/mailer"
	"github.com/O-Nikitin/Social/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		apiURL: env.GetString("EXTERNAL_URL", "localhost:3000"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "hello@demomailtrap.co"),
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", "")},
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", "defaultKey")}},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000")}

	//Logger
	logCfg := zap.NewProductionConfig()
	// Customize time format
	logCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.000000")
	logger := zap.Must(logCfg.Build()).Sugar()
	defer logger.Sync()

	//DB
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxidleTime)
	if err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("DB connected!")
	}
	defer db.Close()
	store := store.NewStorage(db)

	// mailer := mailer.NewSendGridMailer(
	// 	cfg.mail.sendGrid.apiKey,
	// 	cfg.mail.fromEmail)

	mailer, err := mailer.NewMailTrapClient(
		cfg.mail.mailTrap.apiKey,
		cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	app := application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
