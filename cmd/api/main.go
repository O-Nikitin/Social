package main

import (
	"time"

	"github.com/O-Nikitin/Social/internal/auth"
	"github.com/O-Nikitin/Social/internal/db"
	"github.com/O-Nikitin/Social/internal/env"
	"github.com/O-Nikitin/Social/internal/mailer"
	"github.com/O-Nikitin/Social/internal/ratelimiter"
	"github.com/O-Nikitin/Social/internal/store"
	"github.com/O-Nikitin/Social/internal/store/cache"
	"github.com/redis/go-redis/v9"
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
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin")},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "Social",
			}},
		redis: redisConfig{
			address: env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", true)},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		}}

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

	//Redis cache
	var redis *redis.Client
	if cfg.redis.enabled {
		redis = cache.NewRedisClient(
			cfg.redis.address, cfg.redis.pw, cfg.redis.db)
		logger.Info("Redis connected!")
	}
	cacheStorage := cache.NewStorage(redis)

	// mailer := mailer.NewSendGridMailer(
	// 	cfg.mail.sendGrid.apiKey,
	// 	cfg.mail.fromEmail)

	mailer, err := mailer.NewMailTrapClient(
		cfg.mail.mailTrap.apiKey,
		cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	//auth
	auth := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss)

	//Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	app := application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: auth,
		rateLimiter:   rateLimiter,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
