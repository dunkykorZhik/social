package api

import (
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":  "ok",
		"env":     app.Config.Env,
		"version": version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {

		app.internalServerError(w, r, err)
	}
}
