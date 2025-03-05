package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayLoad struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=200"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayLoad struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=200"`
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayLoad	true	"Post payload"
//	@Success		201		{object}	storage.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayLoad
	if err := readJSON(r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}
	post := &storage.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}
	ctx := r.Context()
	if err := app.Storage.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	storage.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.Storage.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return

	}

	ctx := r.Context()
	err = app.Storage.Posts.Delete(ctx, postID)
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

	w.WriteHeader(http.StatusNoContent)

}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayLoad	true	"Post payload"
//	@Success		200		{object}	storage.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdatePostPayLoad
	if err := readJSON(r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}
	post := getPostFromCtx(r)
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	err := app.Storage.Posts.Update(r.Context(), post)
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
	if err := app.CacheStorage.Users.Delete(r.Context(), post.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}

}

func getPostFromCtx(r *http.Request) *storage.Post {
	post := r.Context().Value(postCtx).(*storage.Post)
	return post
}
