package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sergiobonfiglio/tomaccio/internal/config"
	"github.com/sergiobonfiglio/tomaccio/internal/search"
	"github.com/sergiobonfiglio/tomaccio/internal/watched"
)

type stubProvider struct {
	releases []search.Release
	err      error
}

func (p stubProvider) SearchMovie(context.Context, search.MovieSearchQuery) ([]search.Release, error) {
	return p.releases, p.err
}

type stubWatchedProvider struct {
	items []watched.Item
	err   error
}

func (p stubWatchedProvider) ListWatchedMovies(context.Context) ([]watched.Item, error) {
	return p.items, p.err
}

func TestSearchReleasesReturnsPartialResultsAndProviderErrors(t *testing.T) {
	env := &commandEnv{searchProviders: func(*config.Config) ([]search.Provider, []search.ProviderError) {
		return []search.Provider{
			stubProvider{releases: []search.Release{{Provider: "good", Title: "The Matrix 1999", URL: "magnet:?xt=urn:btih:abc", Seeders: 42}}},
			stubProvider{err: errors.New("boom")},
		}, []search.ProviderError{{Provider: "bad-config", Stage: "create", Message: "unsupported"}}
	}}

	got := env.searchReleases(context.Background(), &config.Config{}, search.MovieSearchQuery{Title: "The Matrix", Year: 1999})
	if len(got.Releases) != 1 {
		t.Fatalf("releases=%#v", got.Releases)
	}
	if len(got.Errors) != 2 {
		t.Fatalf("errors=%#v", got.Errors)
	}
	if got.Errors[0].Provider != "bad-config" || got.Errors[1].Stage != "search" || got.Errors[1].Message != "boom" {
		t.Fatalf("errors=%#v", got.Errors)
	}
}

func TestSearchCommandPrintsPartialResultsAndWarnings(t *testing.T) {
	env := &commandEnv{loadConfig: func(string) (*config.Config, error) {
		return &config.Config{Search: config.SearchConfig{Providers: []config.SearchProviderConfig{{Name: "stub", IndexerID: "stub"}}}}, nil
	}, searchProviders: func(*config.Config) ([]search.Provider, []search.ProviderError) {
		return []search.Provider{
			stubProvider{releases: []search.Release{{Provider: "good", Title: "The Matrix 1999 1080p", URL: "magnet:?xt=urn:btih:abc", SizeBytes: 1024 * 1024 * 1024, Seeders: 42}}},
			stubProvider{err: errors.New("boom")},
		}, nil
	}}
	cmd := env.searchCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"The Matrix 1999"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "The Matrix 1999 1080p") {
		t.Fatalf("missing result in %q", text)
	}
	if !strings.Contains(text, "WARN provider error: provider=provider-2 stage=search message=boom") {
		t.Fatalf("missing warning in %q", text)
	}
}

func TestWatchedCommandPrintsJSON(t *testing.T) {
	seen := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	rating := 4.5
	env := &commandEnv{loadConfig: func(string) (*config.Config, error) {
		return &config.Config{Watched: config.WatchedConfig{Plex: config.WatchedPlexConfig{URL: "http://plex", Token: "token"}}}, nil
	}, watchedProvider: func(*config.Config) (watched.Provider, error) {
		return stubWatchedProvider{items: []watched.Item{{Title: "The Matrix", Year: 1999, Rating: &rating, WatchedAt: &seen}}}, nil
	}}

	cmd := env.watchedCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var got []watched.Item
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid json %q: %v", out.String(), err)
	}
	if len(got) != 1 || got[0].Title != "The Matrix" || got[0].Year != 1999 || got[0].Rating == nil || *got[0].Rating != rating {
		t.Fatalf("items=%#v", got)
	}
}
