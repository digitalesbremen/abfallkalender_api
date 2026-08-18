package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReturnsOk(t *testing.T) {
	response := httptest.NewRecorder()

	Health(response, httptest.NewRequest("GET", "/livez", nil))

	if response.Code != http.StatusOK {
		t.Error("Did not get expected HTTP status code, got", response.Code)
	}

	if response.Body.String() != "ok" {
		t.Error("Did not get expected body, got", response.Body.String())
	}

	if response.Header().Get("Content-Type") != "text/plain; charset=UTF-8" {
		t.Error("Did not get expected HTTP content type, got", response.Header().Get("Content-Type"))
	}

	// Probes must never be served from a cache.
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Error("Did not get expected Cache-Control, got", response.Header().Get("Cache-Control"))
	}
}
