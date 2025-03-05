package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dunkykorZhik/social/internal/mailer"
	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterPayLoad struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=70"`
}

type CreateTokenPayLoad struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=70"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterPayLoad	true	"User credentials"
//	@Success		201		{string}	string		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/register [post]
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterPayLoad
	if err := readJSON(r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	user := &storage.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])
	if err := app.Storage.Users.CreateAndInvite(ctx, user, hashToken, app.Config.MailConfig.Exp); err != nil {
		app.internalServerError(w, r, err)

		return
	}

	activationURL := fmt.Sprintf("%s/v1/users/activate/%s", app.Config.Addr, plainToken)
	data := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	status, _ := app.Mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, data)
	/* if err != nil {
		app.Logger.Errorw("error sending welcome email", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err := app.Storage.Users.Delete(ctx, user.ID); err != nil {
			app.Logger.Errorw("error deleting user", "error", err)
		}

		app.internalServerError(w, r, err)
		return

	} */

	app.Logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusAccepted, plainToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateTokenPayLoad	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateTokenPayLoad
	if err := readJSON(r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	user, err := app.Storage.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			app.unAuthError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return

	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unAuthError(w, r, fmt.Errorf("cannot compare the passwords"))
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.Config.AuthConfig.Token.Exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.Config.AuthConfig.Token.Iss,
		"aud": app.Config.AuthConfig.Token.Iss,
	}

	token, err := app.Auth.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
