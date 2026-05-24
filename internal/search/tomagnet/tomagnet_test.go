package tomagnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sergiobonfiglio/tomaccio/internal/search"
)

func TestSearchMovieUsesLocalDefinition(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"The Matrix 1999 1080p BluRay","magnet":"magnet:?xt=urn:btih:abc123","seeders":42,"leechers":7,"size":"2147483648","category":"Movies","published":"2026-05-20T12:00:00Z"}]}`))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	definitionsDir := filepath.Join(tempDir, "definitions")
	if err := os.MkdirAll(definitionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(definitionsDir, "demo.yml"), []byte(`id: demo
name: Demo Indexer
links:
  - https://example.invalid
caps:
  modes:
    search:
      params:
        - name: q
search:
  path: /search
  rows:
    selector: results
  fields:
    title:
      selector: title
    magnet:
      selector: magnet
    seeders:
      selector: seeders
    leechers:
      selector: leechers
    size:
      selector: size
    category:
      selector: category
    date:
      selector: published
`), 0o600); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	client := New("demo-native", "demo", server.URL, 5)
	releases, err := client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "The Matrix", Year: 1999})
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "The Matrix 1999" {
		t.Fatalf("query = %q, want %q", gotQuery, "The Matrix 1999")
	}
	if len(releases) != 1 {
		t.Fatalf("len(releases) = %d, want 1", len(releases))
	}
	if releases[0].Provider != "demo-native" {
		t.Fatalf("provider = %q, want demo-native", releases[0].Provider)
	}
	if releases[0].URL != "magnet:?xt=urn:btih:abc123" {
		t.Fatalf("url = %q", releases[0].URL)
	}
	if releases[0].Seeders != 42 {
		t.Fatalf("seeders = %d, want 42", releases[0].Seeders)
	}
	if releases[0].SizeBytes != 2147483648 {
		t.Fatalf("size = %d, want 2147483648", releases[0].SizeBytes)
	}
	if releases[0].Published != "2026-05-20T12:00:00Z" {
		t.Fatalf("published = %q", releases[0].Published)
	}
}
