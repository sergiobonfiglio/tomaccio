package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Download DownloadConfig `yaml:"download"`
	Search   SearchConfig   `yaml:"search"`
	Watched  WatchedConfig  `yaml:"watched"`
}

type AppConfig struct {
	LogLevel string `yaml:"log_level"`
}

type DownloadConfig struct {
	Label        *string                    `yaml:"label"`
	DirAliases   map[string]string          `yaml:"dir_aliases"`
	Transmission DownloadTransmissionConfig `yaml:"transmission"`
}

type DownloadTransmissionConfig struct {
	URL         string `yaml:"url"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DownloadDir string `yaml:"download_dir"`
}

type SearchConfig struct {
	TimeoutSeconds int                    `yaml:"timeout_seconds"`
	Providers      []SearchProviderConfig `yaml:"providers"`
}

type SearchProviderConfig struct {
	Name           string            `yaml:"name"`
	IndexerID      string            `yaml:"indexer_id"`
	BaseURL        string            `yaml:"base_url"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	Settings       map[string]string `yaml:"settings"`
}

type WatchedConfig struct {
	Plex WatchedPlexConfig `yaml:"plex"`
}

type WatchedPlexConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "./config.yaml"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	expanded, err := expandEnvironment(string(b))
	if err != nil {
		return nil, fmt.Errorf("expand config %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	cfg.ApplyDefaults()
	return &cfg, nil
}

func expandEnvironment(value string) (string, error) {
	missing := map[string]bool{}
	expanded := os.Expand(value, func(name string) string {
		value, ok := os.LookupEnv(name)
		if !ok {
			missing[name] = true
		}
		return value
	})
	if len(missing) == 0 {
		return expanded, nil
	}

	names := make([]string, 0, len(missing))
	for name := range missing {
		names = append(names, name)
	}
	sort.Strings(names)
	return "", fmt.Errorf("unset environment variable(s): %s", strings.Join(names, ", "))
}

func (c *Config) ApplyDefaults() {
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.Download.Label == nil {
		label := "tomaccio"
		c.Download.Label = &label
	}
	if c.Search.TimeoutSeconds <= 0 {
		c.Search.TimeoutSeconds = 120
	}
	for i := range c.Search.Providers {
		if c.Search.Providers[i].Name == "" {
			c.Search.Providers[i].Name = c.Search.Providers[i].IndexerID
		}
	}
}

func (c *Config) Validate(command string) error {
	var missing []string
	switch command {
	case "search":
		validateSearch(c, &missing)
	case "download":
		validateDownload(c, &missing)
	case "watched":
		validateWatched(c, &missing)
	default:
		return errors.New("unknown validation context")
	}
	if len(missing) > 0 {
		return fmt.Errorf("invalid config: missing required field(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateDownload(c *Config, missing *[]string) {
	if c.Download.Transmission.URL == "" {
		*missing = append(*missing, "download.transmission.url")
	}
}

func validateSearch(c *Config, missing *[]string) {
	if len(c.Search.Providers) == 0 {
		return
	}
	for i, p := range c.Search.Providers {
		prefix := fmt.Sprintf("search.providers[%d]", i)
		if p.IndexerID == "" {
			*missing = append(*missing, prefix+".indexer_id")
		}
	}
}

func validateWatched(c *Config, missing *[]string) {
	if c.Watched.Plex.URL == "" {
		*missing = append(*missing, "watched.plex.url")
	}
	if c.Watched.Plex.Token == "" {
		*missing = append(*missing, "watched.plex.token")
	}
}
