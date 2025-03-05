package api

import (
	"net/http"
	"testing"
)

func TestGetUsers(t *testing.T) {
	app := newTestApplication(t)

	mux := app.Mount()

	testToken, _ := app.Auth.GenerateToken(nil)
	t.Run("should not allow unauthenticated users", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := executeRequest(req, mux)
		checkResponse(t, rr.Code, http.StatusUnauthorized)

	})
	t.Run("should allow authenticated users", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)
		rr := executeRequest(req, mux)
		checkResponse(t, rr.Code, http.StatusOK)

	})
}
