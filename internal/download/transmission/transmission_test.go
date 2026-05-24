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
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Method != "torrent-add" {
			t.Fatalf("method = %s", req.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "arguments": map[string]any{"torrent-added": map[string]any{"id": 42}}})
	}))
	defer server.Close()

	client := New(server.URL, "", "", "")
	h, err := client.Add(t.Context(), download.AddDownloadRequest{URL: "magnet:?xt=urn:btih:test"})
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
