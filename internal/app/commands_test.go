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
	"github.com/sergiobonfiglio/tomaccio/internal/definitions"
	"github.com/sergiobonfiglio/tomaccio/internal/download"
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

type stubDownloader struct {
	handle *download.DownloadHandle
	addReq download.AddDownloadRequest
	addErr error
}

func (d *stubDownloader) Test(context.Context) error { return nil }
func (d *stubDownloader) Add(_ context.Context, req download.AddDownloadRequest) (*download.DownloadHandle, error) {
	d.addReq = req
	if d.addErr != nil {
		return nil, d.addErr
	}
	if d.handle != nil {
		return d.handle, nil
	}
	return &download.DownloadHandle{Provider: "transmission", ID: "1"}, nil
}
func (d *stubDownloader) List(context.Context) ([]download.DownloadStatus, error) { return nil, nil }
func (d *stubDownloader) Get(context.Context, download.DownloadHandle) (*download.DownloadStatus, error) {
	return nil, nil
}
func (d *stubDownloader) Cancel(context.Context, download.DownloadHandle) error { return nil }
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

func TestDefinitionsSyncCommandPrintsSyncedCount(t *testing.T) {
	env := &commandEnv{definitionsSync: func() (definitions.Metadata, error) {
		return definitions.Metadata{Files: []string{"yts.yml", "1337x.yml"}}, nil
	}}
	cmd := env.definitionsCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"sync"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "synced 2 definitions to .tomaccio/definitions") {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestDownloadAddCommandPassesDirOverride(t *testing.T) {
	dl := &stubDownloader{handle: &download.DownloadHandle{Provider: "transmission", ID: "42"}}
	env := &commandEnv{
		loadConfig: func(string) (*config.Config, error) {
			return &config.Config{Download: config.DownloadConfig{Transmission: config.DownloadTransmissionConfig{URL: "http://transmission"}}}, nil
		},
		downloader: func(*config.Config) (download.Downloader, error) { return dl, nil },
	}
	cmd := env.downloadCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"add", "--dir", "/media/movies", "magnet:?xt=urn:btih:abc"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if dl.addReq.URL != "magnet:?xt=urn:btih:abc" {
		t.Fatalf("url = %q", dl.addReq.URL)
	}
	if dl.addReq.DownloadDir != "/media/movies" {
		t.Fatalf("download dir = %q", dl.addReq.DownloadDir)
	}
	if !strings.Contains(out.String(), "Added transmission:42") {
		t.Fatalf("unexpected output %q", out.String())
	}
}
