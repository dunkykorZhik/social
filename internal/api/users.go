package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// GetUserHandler godoc
//
//	@Summary		Fetches the User
//	@Description	Fetches the User info using ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	storage.User
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
func (app *Application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return

	}

	user, err := app.getUser(r.Context(), userID)

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

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowTheUser godoc
//
//	@Summary		Follow The User
//	@Description	Gets the id of User and the id of User to follow
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		202		{string}	string	"User Followed"
//	@Failure		409		{object}	error	"Cannot Follow That User"
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	fUser, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := app.Storage.Users.Follow(r.Context(), fUser, user.ID); err != nil {
		switch err {
		case storage.ErrConflict:
			app.conflictError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		202		{string}	string	"User Followed"
//	@Failure		409		{object}	error	"Cannot UnFollow That User"
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *Application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	fUser, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := app.Storage.Users.UnFollow(r.Context(), fUser, user.ID); err != nil {
		switch err {
		case storage.ErrConflict:
			app.conflictError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}

	}

	w.WriteHeader(http.StatusAccepted)
}

func getUserFromCtx(r *http.Request) *storage.User {
	user, ok := r.Context().Value(userCtx).(*storage.User)
	if !ok {
		return nil
	}
	return user
}

// getUserFeed godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]storage.PostForFeed
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *Application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	pq := storage.PaginateQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Search: "",
		Tags:   []string{},
	}

	pq, err := pq.Parse(r)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	err = Validate.Struct(pq)
	if err != nil {
		app.badRequestReponse(w, r, err)
	}
	user := getUserFromCtx(r)

	ctx := r.Context()
	feed, err := app.Storage.Users.GetUserFeed(ctx, user.ID, pq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err = app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)

	}
}

// activateUser godoc
//
//	@Summary		Activates the user using token from invitation
//	@Description	Activates the user using token from invitation
//	@Tags			auth
//	@Produce		json
//	@Param			token	path		string	true	"token"
//	@Success		202		{string}	string	"User Activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := app.Storage.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case storage.ErrNotFound:
			app.notFoundReponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusAccepted, ""); err != nil {
		app.internalServerError(w, r, err)
	}

}
