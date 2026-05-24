package app

import (
	"github.com/sergiobonfiglio/tomaccio/internal/config"
	"github.com/sergiobonfiglio/tomaccio/internal/download"
	"github.com/sergiobonfiglio/tomaccio/internal/download/transmission"
	"github.com/sergiobonfiglio/tomaccio/internal/search"
	"github.com/sergiobonfiglio/tomaccio/internal/search/tomagnet"
	"github.com/sergiobonfiglio/tomaccio/internal/watched"
	"github.com/sergiobonfiglio/tomaccio/internal/watched/plex"
)

func newDownloader(cfg *config.Config) (download.Downloader, error) {
	t := cfg.Download.Transmission
	return transmission.New(t.URL, t.Username, t.Password, t.DownloadDir), nil
}

func newSearchProviders(cfg *config.Config) ([]search.Provider, []search.ProviderError) {
	providers := make([]search.Provider, 0, len(cfg.Search.Providers))
	for _, p := range cfg.Search.Providers {
		providers = append(providers, tomagnet.New(p.Name, p.IndexerID, p.BaseURL, p.TimeoutSeconds))
	}
	return providers, nil
}

func newWatchedProvider(cfg *config.Config) (watched.Provider, error) {
	p := cfg.Watched.Plex
	return plex.New(p.URL, p.Token), nil
}
