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
	label := ""
	if cfg.Download.Label != nil {
		label = *cfg.Download.Label
	}
	return transmission.New(t.URL, t.Username, t.Password, t.DownloadDir, label), nil
}

func newSearchProviders(cfg *config.Config) ([]search.Provider, []search.ProviderError) {
	if len(cfg.Search.Providers) == 0 {
		return []search.Provider{
			tomagnet.New("yts", "yts", "auto", 0),
			tomagnet.New("1337x", "1337x", "auto", 0),
			tomagnet.New("thepiratebay", "thepiratebay", "auto", 0),
		}, nil
	}

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
