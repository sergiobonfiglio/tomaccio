package tomagnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergiobonfiglio/tomaccio/internal/search"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

func TestSearchMovieUsesLocalDefinition(t *testing.T) {
	var gotQuery string
	var gotType string
	var gotTitle string
	var gotYear string
	var gotTMDBID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		gotType = r.URL.Query().Get("t")
		gotTitle = r.URL.Query().Get("title")
		gotYear = r.URL.Query().Get("year")
		gotTMDBID = r.URL.Query().Get("tmdbid")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"The Matrix 1999 1080p BluRay","magnet":"magnet:?xt=urn:btih:abc123","seeders":42,"leechers":7,"size":"2147483648","category":"Movies","published":"2026-05-20T12:00:00Z"}]}`))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	definitionsDir := filepath.Join(tempDir, ".tomaccio", "definitions")
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
  inputs:
    q: "{{ .Keywords }}"
    t: "{{ .Query.Type }}"
    title: "{{ .Query.Title }}"
    year: "{{ .Query.Year }}"
    tmdbid: '{{ if ne .Query.TMDBID .False }}{{ .Query.TMDBID }}{{ end }}'
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

	client := New("demo-native", "demo", server.URL, 5, nil)
	releases, err := client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "The Matrix", Year: 1999})
	if err != nil {
		t.Fatal(err)
	}
	if gotType != "search" {
		t.Fatalf("type = %q, want %q", gotType, "search")
	}
	if gotQuery != "The Matrix 1999" {
		t.Fatalf("query = %q, want %q", gotQuery, "The Matrix 1999")
	}
	if gotTitle != "The Matrix" {
		t.Fatalf("title = %q, want %q", gotTitle, "The Matrix")
	}
	if gotYear != "1999" {
		t.Fatalf("year = %q, want %q", gotYear, "1999")
	}
	if gotTMDBID != "" {
		t.Fatalf("tmdbid = %q, want empty", gotTMDBID)
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

func TestSearchMovieAuthenticatesWithDefinitionSettings(t *testing.T) {
	const (
		username = "fixture-user"
		password = "fixture-password"
		cookie   = "fixture-session"
	)
	var gotUsername, gotPassword string
	var setLoginCookie, loginTestGotCookie, searchGotCookie bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/login":
			_, _ = w.Write([]byte(`<form action="/session"><input name="username"><input name="password"></form>`))
		case r.Method == http.MethodPost && r.URL.Path == "/session":
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			gotUsername = r.Form.Get("username")
			gotPassword = r.Form.Get("password")
			if gotUsername == username && gotPassword == password {
				http.SetCookie(w, &http.Cookie{Name: "session", Value: cookie, Path: "/"})
				setLoginCookie = true
			}
		case r.Method == http.MethodGet && r.URL.Path == "/login-test":
			if c, err := r.Cookie("session"); err == nil && c.Value == cookie {
				loginTestGotCookie = true
				_, _ = w.Write([]byte(`<div class="authenticated"></div>`))
				return
			}
			_, _ = w.Write([]byte(`<div class="signed-out"></div>`))
		case r.Method == http.MethodGet && r.URL.Path == "/search":
			if c, err := r.Cookie("session"); err == nil && c.Value == cookie {
				searchGotCookie = true
			} else {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"title":"Authenticated Result","magnet":"magnet:?xt=urn:btih:authenticated"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	useTempDefinition(t, "authenticated", `id: authenticated
name: Authenticated Indexer
links:
  - https://example.invalid
settings:
  - name: username
  - name: password
caps:
  modes:
    search: [q]
login:
  method: form
  path: /login
  form: form
  inputs:
    username: "{{ .Config.username }}"
    password: "{{ .Config.password }}"
  test:
    path: /login-test
    selector: .authenticated
search:
  path: /search
  inputs:
    q: "{{ .Keywords }}"
  rows:
    selector: results
  fields:
    title:
      selector: title
    magnet:
      selector: magnet
`)

	client := New("authenticated", "authenticated", server.URL, 5, map[string]string{
		"username": username,
		"password": password,
	})
	releases, err := client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "Dune"})
	if err != nil {
		t.Fatal(err)
	}
	if gotUsername != username || gotPassword != password {
		t.Fatalf("login credentials = %q/%q", gotUsername, gotPassword)
	}
	if !setLoginCookie || !loginTestGotCookie || !searchGotCookie {
		t.Fatalf("cookie flow: set=%t login_test=%t search=%t", setLoginCookie, loginTestGotCookie, searchGotCookie)
	}
	if len(releases) != 1 || releases[0].Title != "Authenticated Result" {
		t.Fatalf("releases = %#v", releases)
	}

	wrongPassword := "wrong-fixture-password"
	client = New("authenticated", "authenticated", server.URL, 5, map[string]string{
		"username": username,
		"password": wrongPassword,
	})
	_, err = client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "Dune"})
	if err == nil || !strings.Contains(err.Error(), "login test failed") {
		t.Fatalf("error = %v, want authentication error", err)
	}
	if strings.Contains(err.Error(), wrongPassword) {
		t.Fatalf("error contains password: %q", err)
	}
}

func TestSearchMovieIncludesResultEnrichmentError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"title":"Legacy Result","details":"/topic/1"}]}`))
		case "/topic/1":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	useTempDefinition(t, "legacy", `id: legacy
name: Legacy Indexer
links:
  - https://example.invalid
caps:
  modes:
    search: [q]
search:
  path: /search
  inputs:
    q: "{{ .Keywords }}"
  rows:
    selector: results
  fields:
    title:
      selector: title
    details:
      selector: details
    download:
      selector: download
download:
  before:
    pathselector:
      selector: a.thanks
      attribute: href
  selectors:
    - selector: a[href^="magnet:"]
      attribute: href
`)

	client := New("legacy", "legacy", server.URL, 5, nil)
	releases, err := client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "Legacy"})
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 1 || releases[0].ResolutionError != "download before path not found" {
		t.Fatalf("releases = %#v", releases)
	}
	if releases[0].URL != server.URL+"/topic/1" {
		t.Fatalf("release URL = %q", releases[0].URL)
	}
}

func TestApplyDefinitionSettingsRejectsUnknownSetting(t *testing.T) {
	definition := &tomagnetlib.Definition{
		ID:       "private",
		Settings: []tomagnetlib.Setting{{Name: "username"}},
	}

	err := applyDefinitionSettings(definition, map[string]string{"usernmae": "fixture-user"})
	if err == nil || !strings.Contains(err.Error(), `indexer "private" does not declare setting "usernmae"`) {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "fixture-user") {
		t.Fatalf("error contains setting value: %q", err)
	}
}

func TestNewClonesSettings(t *testing.T) {
	settings := map[string]string{"username": "original"}
	client := New("private", "private", "auto", 5, settings)
	settings["username"] = "changed"

	if client.settings["username"] != "original" {
		t.Fatalf("client settings = %#v", client.settings)
	}
}

func TestSearchMovieRedactsSettingsFromErrors(t *testing.T) {
	const apiKey = "fixture-api-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "failed", http.StatusInternalServerError)
	}))
	defer server.Close()

	useTempDefinition(t, "api", `id: api
name: API Indexer
links:
  - https://example.invalid
settings:
  - name: api_key
caps:
  modes:
    search: [q]
search:
  path: "/search?api_key={{ .Config.api_key }}"
  rows:
    selector: results
  fields:
    title:
      selector: title
    magnet:
      selector: magnet
`)

	client := New("api", "api", server.URL, 5, map[string]string{"api_key": apiKey})
	_, err := client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "Dune"})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), apiKey) || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("error = %q", err)
	}
}

func TestSearchMovieReportsMissingDefinition(t *testing.T) {
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	}()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	client := New("missing", "missing", "auto", 5, nil)
	_, err = client.SearchMovie(context.Background(), search.MovieSearchQuery{Title: "Dune"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "definitions sync") {
		t.Fatalf("error = %q", err.Error())
	}
}

func useTempDefinition(t *testing.T, id, definition string) {
	t.Helper()
	tempDir := t.TempDir()
	definitionsDir := filepath.Join(tempDir, ".tomaccio", "definitions")
	if err := os.MkdirAll(definitionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(definitionsDir, id+".yml"), []byte(definition), 0o600); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Errorf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
}
