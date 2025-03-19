package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRestApiClient(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	handler.HandleFunc("/created", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler.HandleFunc("/nocontent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"type": "bad_request", "message": "Invalid request"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := server.Client()

	apiClient:= NewRestApiClient(client, baseURL, "testuser", "testpassword")
	t.Run("Test GET Success", func(t *testing.T) {
		resp, err := apiClient.GetRequest("/success")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Test POST Created", func(t *testing.T) {
		resp, err := apiClient.PostRequest("/created", []byte(`{"key": "value"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}
	})

	t.Run("Test PUT No Content", func(t *testing.T) {
		resp, err := apiClient.PutRequest("/nocontent", []byte(`{"key": "value"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}
	})

	t.Run("Test DELETE Error", func(t *testing.T) {
		_, err := apiClient.DeleteRequest("/error")
		if err == nil || !strings.Contains(err.Error(), "Invalid request") {
			t.Errorf("expected error containing 'Invalid request', got %v", err)
		}
	})
}
