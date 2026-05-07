package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthReturnsOK(t *testing.T) {
	fixedTime := time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC)
	handler := NewRouter(Options{Now: func() time.Time { return fixedTime }})

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", contentType)
	}

	var body healthResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status = %q, want ok", body.Status)
	}
	if body.Time != "2026-05-06T01:02:03Z" {
		t.Fatalf("time = %q, want fixed time", body.Time)
	}
}

func TestHealthIncludesDatabaseStatus(t *testing.T) {
	handler := NewRouter(Options{
		HealthCheck: func(_ context.Context) error {
			return nil
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	var body healthResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Database != "ok" {
		t.Fatalf("database = %q, want ok", body.Database)
	}
}

func TestHealthzReturnsOK(t *testing.T) {
	handler := NewRouter(Options{
		HealthCheck: func(_ context.Context) error {
			return nil
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	var body healthzResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.OK {
		t.Fatal("ok = false, want true")
	}
}
