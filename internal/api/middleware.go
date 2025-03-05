package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

/* func (app *Application) userContextMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestReponse(w, r, err)
			return

		}

		ctx := r.Context()
		user, err := app.Storage.Users.GetByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, storage.ErrNotFound):
				app.notFoundReponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return

			}

		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})

} */

func (app *Application) postContextMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
		if err != nil {
			app.badRequestReponse(w, r, err)
			return

		}

		ctx := r.Context()
		post, err := app.Storage.Posts.GetByID(ctx, postID)
		if err != nil {
			switch {
			case errors.Is(err, storage.ErrNotFound):
				app.notFoundReponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return

			}

		}
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) basicAuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if auth == "" {
			app.basicUnAuthError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			app.basicUnAuthError(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			app.basicUnAuthError(w, r, err)
			return
		}
		username := app.Config.AuthConfig.Basic.Username
		password := app.Config.AuthConfig.Basic.Password
		credentials := strings.Split(string(decoded), ":")
		if len(credentials) != 2 || credentials[0] != username || credentials[1] != password {
			app.basicUnAuthError(w, r, fmt.Errorf("invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)

	})
}
func (app *Application) RateLimiterMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.RateLimiterConfig.Enabled {
			if allow, retryAfter := app.RateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})

}

func (app *Application) authMaiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			app.unAuthError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unAuthError(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token, err := app.Auth.ValidateToken(parts[1])
		if err != nil {
			app.unAuthError(w, r, fmt.Errorf("authorization header is malformed"))
			return

		}
		claims, _ := token.Claims.(jwt.MapClaims)

		userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
		}
		ctx := r.Context()
		user, err := app.getUser(ctx, userId)
		if err != nil {
			app.internalServerError(w, r, err)
		}

		ctx = context.WithValue(ctx, userCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (app *Application) checkPermission(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post := getPostFromCtx(r)
		app.Logger.Infof("the post id %v", post.ID)
		user := getUserFromCtx(r)
		app.Logger.Infof("the post id %v", post.ID)
		if user == nil || post == nil {
			app.internalServerError(w, r, fmt.Errorf("cannot get user or post from context"))
			return
		}
		//TODO : fix userInvitation
		role, err := app.Storage.Roles.GetByName(r.Context(), requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if post.UserID != user.ID && user.Role_id < role.ID {
			app.forbiddenResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) getUser(ctx context.Context, userID int64) (*storage.User, error) {
	if !app.Config.RedisConfig.Enabled {
		return app.Storage.Users.GetByID(ctx, userID)

	}
	user, err := app.CacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {

		user, err = app.Storage.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.CacheStorage.Users.Set(ctx, user); err != nil {

			return nil, err
		}
	}

	return user, nil
}
