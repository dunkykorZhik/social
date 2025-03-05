package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dunkykorZhik/social/docs"
	"github.com/dunkykorZhik/social/internal/auth"
	"github.com/dunkykorZhik/social/internal/env"
	"github.com/dunkykorZhik/social/internal/mailer"
	"github.com/dunkykorZhik/social/internal/rateLimiter"
	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/dunkykorZhik/social/internal/storage/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	//httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
)

const version = "0.0.1"

type Application struct {
	Config       Config
	Storage      storage.Storage
	CacheStorage cache.Storage
	Logger       *zap.SugaredLogger
	Mailer       mailer.Client
	Auth         auth.Authenticator
	RateLimiter  rateLimiter.RateLimiter
}

type Config struct {
	Addr              string
	Db                DbConfig
	Env               string
	ExternalAddr      string
	MailConfig        MailConfig
	AuthConfig        AuthConfig
	RedisConfig       RedisConfig
	RateLimiterConfig rateLimiter.Config
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

type AuthConfig struct {
	Basic BasicConfig
	Token TokenConfig
}

type TokenConfig struct {
	Secret string
	Exp    time.Duration
	Iss    string
}

type BasicConfig struct {
	Username string
	Password string
}

type MailConfig struct {
	FromEmail string
	ApiKey    string
	Exp       time.Duration
}
type DbConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func (app *Application) Mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.RateLimiterMiddleWare)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.With(app.basicAuthMiddleWare).Get("/health", app.healthCheckHandler)

		sUrl := fmt.Sprintf("%s/swagger/doc.json", app.Config.Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(sUrl)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.authMaiddleWare)
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleWare)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPermission("moderator", app.deletePostHandler))
				r.Patch("/", app.checkPermission("admin", app.updatePostHandler))
			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.authMaiddleWare)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.authMaiddleWare)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})
	return r
}

func (app *Application) Run(h *http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.Config.ExternalAddr
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      *h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.Logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.Logger.Infow("server has stopped", "addr", app.Config.Addr, "env", app.Config.Env)

	return nil
	//TODO read about GraceFul shutdown
}
