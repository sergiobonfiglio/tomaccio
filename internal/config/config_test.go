package config

import (
	"strings"
	"testing"
)

func TestValidateWatchedRequiresPlexConfig(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate("watched")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "watched.plex.url") || !strings.Contains(msg, "watched.plex.token") {
		t.Fatalf("error=%q", msg)
	}
}

func TestValidateWatchedAcceptsPlexConfig(t *testing.T) {
	cfg := &Config{Watched: WatchedConfig{Plex: WatchedPlexConfig{URL: "http://plex", Token: "token"}}}
	if err := cfg.Validate("watched"); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateSearchAllowsDefaultPublicProviders(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate("search"); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateSearchRequiresIndexerIDButNotName(t *testing.T) {
	cfg := &Config{Search: SearchConfig{Providers: []SearchProviderConfig{{IndexerID: "thepiratebay"}}}}
	if err := cfg.Validate("search"); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestApplyDefaultsSetsDefaultDownloadLabel(t *testing.T) {
	cfg := &Config{}
	cfg.ApplyDefaults()
	if cfg.Download.Label == nil || *cfg.Download.Label != "tomaccio" {
		t.Fatalf("label = %#v", cfg.Download.Label)
	}
}

func TestApplyDefaultsSetsSearchProviderNameFromIndexerID(t *testing.T) {
	cfg := &Config{Search: SearchConfig{Providers: []SearchProviderConfig{{IndexerID: "thepiratebay"}}}}
	cfg.ApplyDefaults()
	if got := cfg.Search.Providers[0].Name; got != "thepiratebay" {
		t.Fatalf("provider name = %q", got)
	}
}

func TestApplyDefaultsPreservesExplicitEmptyDownloadLabel(t *testing.T) {
	empty := ""
	cfg := &Config{Download: DownloadConfig{Label: &empty}}
	cfg.ApplyDefaults()
	if cfg.Download.Label == nil || *cfg.Download.Label != "" {
		t.Fatalf("label = %#v", cfg.Download.Label)
	}
}
