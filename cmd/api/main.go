package main

import (
	"path/filepath"
	"time"

	"github.com/dunkykorZhik/social/internal/api"
	"github.com/dunkykorZhik/social/internal/auth"
	"github.com/dunkykorZhik/social/internal/db"
	env "github.com/dunkykorZhik/social/internal/env"
	"github.com/dunkykorZhik/social/internal/mailer"
	"github.com/dunkykorZhik/social/internal/rateLimiter"
	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/dunkykorZhik/social/internal/storage/cache"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	// http-swagger middleware
)

// @title Social Documentation

// @description Project from Tiego`s Go Backend Course
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//@securityDefinitions.apikey	ApiKeyAuth
//@in header
//@name Authorization
//@description dummy dum

func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	rootDir, err := filepath.Abs(".")
	if err != nil {
		logger.Fatalf("Error getting root directory: %v", err)
	}

	// Load .env file explicitly
	envPath := filepath.Join(rootDir, ".env")
	err = godotenv.Load(envPath)

	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	cfg := api.Config{
		Addr: env.GetString("ADDR", ":4040"),
		Db: api.DbConfig{
			Addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:8041/social?sslmode=disable"),
			MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		Env:          env.GetString("ENV", "development"),
		ExternalAddr: env.GetString("EXT_ADDR", "localhost:4040"),
		MailConfig: api.MailConfig{
			FromEmail: env.GetString("FROM_EMAIL", "korkemay.oserbay@nu.edu.kz"),
			ApiKey:    env.GetString("MAILTRAP_API_KEY", "6fd0d97cdbdffec4ade669008f6cb1dd"),
			Exp:       time.Hour * 24 * 3,
		},
		AuthConfig: api.AuthConfig{
			Basic: api.BasicConfig{
				Username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				Password: env.GetString("BASIC_AUTH_PASS", "admin"),
			},
			Token: api.TokenConfig{
				Secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				Exp:    time.Hour * 24 * 3, // 3 days
				Iss:    "gophersocial",
			},
		},
		RedisConfig: api.RedisConfig{
			Addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			Password: env.GetString("REDIS_PW", ""),
			DB:       env.GetInt("REDIS_DB", 0),
			Enabled:  env.GetBool("REDIS_ENABLED", false),
		},
		RateLimiterConfig: rateLimiter.Config{
			RequestPerTF: env.GetInt("RATE_LIMITER_RPTF", 20),
			TimeFrame:    time.Second * 5,
			Enabled:      env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	db, err := db.New(cfg.Db.Addr, cfg.Db.MaxOpenConns, cfg.Db.MaxIdleConns, cfg.Db.MaxIdleTime)

	//err := db.New("postgres://admin:adminpassword@localhost/social?sslmode=disable", 3, 3, "15m")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	str := storage.NewStorage(db)
	logger.Infow("Db connection success")

	mailer, err := mailer.NewMailTrapClient(cfg.MailConfig.ApiKey, cfg.MailConfig.FromEmail)
	if err != nil {
		logger.Fatal(err)

	}

	auth := auth.NewAuthenticator(cfg.AuthConfig.Token.Secret, cfg.AuthConfig.Token.Iss, cfg.AuthConfig.Token.Iss)

	var rdb *redis.Client
	if cfg.RedisConfig.Enabled {
		rdb = cache.NewRedisClient(cfg.RedisConfig.Addr, cfg.RedisConfig.Password, cfg.RedisConfig.DB)
		logger.Infow("Redis connection success")
		defer rdb.Close()
	}

	cacheStr := cache.NewRedisStorage(rdb)

	rateL := rateLimiter.NewRateLimiter(cfg.RateLimiterConfig.RequestPerTF, cfg.RateLimiterConfig.TimeFrame)
	app := &api.Application{
		Config:       cfg,
		Storage:      str,
		CacheStorage: cacheStr,
		Logger:       logger,
		Mailer:       mailer,
		Auth:         auth,
		RateLimiter:  rateL,
	}

	mux := app.Mount()
	logger.Fatal(app.Run(&mux))

}

/*

DB_ADDR=postgres://admin:adminpassword@localhost/social?sslmode:disable
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=30
DB_MAX_IDLE_TIME=15m
*/

//ToDO learn about mailtrap
