package transmission

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergiobonfiglio/tomaccio/internal/download"
)

func TestClientNegotiatesSessionAndAddsTorrent(t *testing.T) {
	var sawSession bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(sessionHeader) == "" {
			w.Header().Set(sessionHeader, "sid-1")
			w.WriteHeader(http.StatusConflict)
			return
		}
		sawSession = true
		var req struct {
			Method    string         `json:"method"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Method != "torrent-add" {
			t.Fatalf("method = %s", req.Method)
		}
		if req.Arguments["download-dir"] != "/media/movies" {
			t.Fatalf("download-dir = %#v", req.Arguments["download-dir"])
		}
		labels, ok := req.Arguments["labels"].([]any)
		if !ok || len(labels) != 1 || labels[0] != "tomaccio" {
			t.Fatalf("labels = %#v", req.Arguments["labels"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "arguments": map[string]any{"torrent-added": map[string]any{"id": 42}}})
	}))
	defer server.Close()

	client := New(server.URL, "", "", "", "tomaccio")
	h, err := client.Add(t.Context(), download.AddDownloadRequest{URL: "magnet:?xt=urn:btih:test", DownloadDir: "/media/movies"})
	if err != nil {
		t.Fatal(err)
	}
	if !sawSession {
		t.Fatal("expected session header retry")
	}
	if h.Provider != "transmission" || h.ID != "42" {
		t.Fatalf("unexpected handle: %#v", h)
	}
}

func TestClientListIncludesLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(sessionHeader) == "" {
			w.Header().Set(sessionHeader, "sid-1")
			w.WriteHeader(http.StatusConflict)
			return
		}
		var req struct {
			Method    string         `json:"method"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Method != "torrent-get" {
			t.Fatalf("method = %s", req.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "arguments": map[string]any{"torrents": []map[string]any{{"id": 42, "name": "The Matrix", "percentDone": 0.5, "status": 4, "labels": []string{"tomaccio", "movies"}}}}})
	}))
	defer server.Close()

	client := New(server.URL, "", "", "", "tomaccio")
	items, err := client.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %#v", items)
	}
	if len(items[0].Labels) != 2 || items[0].Labels[0] != "tomaccio" || items[0].Labels[1] != "movies" {
		t.Fatalf("labels = %#v", items[0].Labels)
	}
}
