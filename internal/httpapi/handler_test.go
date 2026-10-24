package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shorter/internal/link"
	"url-shorter/internal/storage/memory"
)

func testRouter() http.Handler {
	return NewRouter(NewHandler(link.NewService(memory.NewStore()), "http://localhost:8080"))
}

func TestHealthAndInfo(t *testing.T) {
	router := testRouter()
	for _, tt := range []struct {
		path string
		body string
	}{
		{"/health", `{"status":"ok"}`},
		{"/api/info", `{"name":"URL Shortener","version":"1.0.0"}`},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != tt.body {
			t.Fatalf("GET %s = %d %q", tt.path, response.Code, response.Body.String())
		}
	}
}

func TestCreateAndRedirect(t *testing.T) {
	router := testRouter()
	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/link", strings.NewReader(`{"url":"https://example.com/path"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /api/link = %d: %s", created.Code, created.Body.String())
	}

	var body LinkResponse
	if err := json.NewDecoder(created.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code == "" || body.ShortURL != "http://localhost:8080/"+body.Code || body.OriginalURL != "https://example.com/path" {
		t.Fatalf("unexpected response: %+v", body)
	}

	redirect := httptest.NewRecorder()
	router.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, "/"+body.Code, nil))
	if redirect.Code != http.StatusFound || redirect.Header().Get("Location") != body.OriginalURL {
		t.Fatalf("GET /%s = %d, Location %q", body.Code, redirect.Code, redirect.Header().Get("Location"))
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	router := testRouter()
	for _, payload := range []string{`{`, `{"url":"test"}`, `{"url":"ftp://example.com"}`} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/link", strings.NewReader(payload)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("POST %q = %d", payload, response.Code)
		}
	}
}

func TestGetMissingLink(t *testing.T) {
	response := httptest.NewRecorder()
	testRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /missing = %d", response.Code)
	}
}

type failingRepository struct{}

func (failingRepository) Save(context.Context, link.Link) error {
	return errors.New("storage unavailable")
}
func (failingRepository) Get(context.Context, string) (link.Link, error) {
	return link.Link{}, errors.New("storage unavailable")
}

func TestStorageFailureReturnsServerError(t *testing.T) {
	router := NewRouter(NewHandler(link.NewService(failingRepository{}), "http://localhost:8080"))
	for _, tt := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/link", `{"url":"https://example.com"}`},
		{http.MethodGet, "/code", ""},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("%s %s = %d", tt.method, tt.path, response.Code)
		}
	}
}
