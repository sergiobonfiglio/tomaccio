package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSearchProviderSettingsExpandsEnvironmentAndScalarValues(t *testing.T) {
	t.Setenv("TEST_INDEXER_USERNAME", "test-user")
	t.Setenv("TEST_INDEXER_PASSWORD", "test-password")

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`search:
  timeout_seconds: 240
  providers:
    - indexer_id: private
      settings:
        username: "${TEST_INDEXER_USERNAME}"
        password: "${TEST_INDEXER_PASSWORD}"
        freeleech: false
        api_key: test-api-key
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Search.TimeoutSeconds != 240 {
		t.Fatalf("search timeout = %d, want 240", cfg.Search.TimeoutSeconds)
	}
	got := cfg.Search.Providers[0].Settings
	want := map[string]string{
		"username":  "test-user",
		"password":  "test-password",
		"freeleech": "false",
		"api_key":   "test-api-key",
	}
	for name, value := range want {
		if got[name] != value {
			t.Errorf("settings[%q] = %q, want %q", name, got[name], value)
		}
	}
}

func TestLoadSearchProviderWithoutSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("search:\n  providers:\n    - indexer_id: thepiratebay\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Search.Providers[0].Settings != nil {
		t.Fatalf("settings = %#v, want nil", cfg.Search.Providers[0].Settings)
	}
}

func TestLoadReportsUnsetEnvironmentVariable(t *testing.T) {
	const name = "TOMACCIO_TEST_UNSET_INDEXER_PASSWORD"
	old, wasSet := os.LookupEnv(name)
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(name, old)
		} else {
			_ = os.Unsetenv(name)
		}
	})

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("search:\n  providers:\n    - indexer_id: private\n      settings:\n        password: ${"+name+"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), name) || strings.Contains(err.Error(), "password:") {
		t.Fatalf("error = %q", err)
	}
}

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

func TestApplyDefaultsSetsSearchTimeout(t *testing.T) {
	cfg := &Config{}
	cfg.ApplyDefaults()
	if cfg.Search.TimeoutSeconds != 120 {
		t.Fatalf("search timeout = %d, want 120", cfg.Search.TimeoutSeconds)
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
