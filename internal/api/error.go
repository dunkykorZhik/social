package api

import (
	"net/http"
)

func (app *Application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}
func (app *Application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusConflict, "the conflict with resources")
}

func (app *Application) badRequestReponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}
func (app *Application) notFoundReponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnf("response not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *Application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.Logger.Warnf("forbidden", "method", r.Method, "path", r.URL.Path)
	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (app *Application) basicUnAuthError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("unauthorized basic auth error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) unAuthError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("unauthorized basic auth error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	//w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.Logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
