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
