package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dunkykorZhik/social/internal/auth"
	"github.com/dunkykorZhik/social/internal/storage"
	"github.com/dunkykorZhik/social/internal/storage/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()
	logger := zap.NewNop().Sugar()
	mockStorage := storage.NewMockStorage()
	mockCacheStorage := cache.NewMockStorage()

	mockAuth := auth.NewMockAuthenticator()

	return &Application{
		Logger:       logger,
		Storage:      mockStorage,
		CacheStorage: mockCacheStorage,
		Auth:         mockAuth,
	}

}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponse(t *testing.T, res, expected int) {
	if expected != res {
		t.Errorf("Expected %d, but got %d", expected, res)
	}

}
