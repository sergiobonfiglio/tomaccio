package plex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListWatchedMovies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("X-Plex-Token") != "token" && r.Header.Get("X-Plex-Token") != "token" {
			t.Fatalf("missing token on %s", r.URL.String())
		}
		switch r.URL.Path {
		case "/library/sections":
			w.Write([]byte(`<MediaContainer><Directory key="1" type="movie"/><Directory key="2" type="show"/></MediaContainer>`))
		case "/library/sections/1/all":
			if r.URL.Query().Get("viewed") != "1" {
				t.Fatalf("viewed query=%q", r.URL.RawQuery)
			}
			w.Write([]byte(`<MediaContainer><Video title="The Matrix" year="1999" userRating="4.5" lastViewedAt="1704207845" viewCount="1"/></MediaContainer>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	items, err := New(server.URL, "token").ListWatchedMovies(context.Background())
	if err != nil {
		t.Fatalf("ListWatchedMovies() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%#v", items)
	}
	if items[0].Title != "The Matrix" || items[0].Year != 1999 || items[0].Rating == nil || *items[0].Rating != 4.5 {
		t.Fatalf("item=%#v", items[0])
	}
	want := time.Unix(1704207845, 0).UTC()
	if items[0].WatchedAt == nil || !items[0].WatchedAt.Equal(want) {
		t.Fatalf("watched_at=%v want %v", items[0].WatchedAt, want)
	}
}
