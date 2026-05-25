package app

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sergiobonfiglio/tomaccio/internal/config"
	"github.com/sergiobonfiglio/tomaccio/internal/download"
	"github.com/sergiobonfiglio/tomaccio/internal/search"
	"github.com/sergiobonfiglio/tomaccio/internal/watched"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

func (e *commandEnv) downloadCommand() *cobra.Command {
	dl := &cobra.Command{Use: "download", Short: "Downloader commands"}
	dl.AddCommand(&cobra.Command{Use: "check", Short: "Check downloader connectivity", RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := e.load("download")
		if err != nil {
			return err
		}
		d, err := e.newDownloader(cfg)
		if err != nil {
			return err
		}
		if err := d.Test(cmd.Context()); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Downloader OK")
		return nil
	}})
	var listLabel string
	list := &cobra.Command{Use: "list", Short: "List downloads", RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := e.load("download")
		if err != nil {
			return err
		}
		d, err := e.newDownloader(cfg)
		if err != nil {
			return err
		}
		items, err := d.List(cmd.Context())
		if err != nil {
			return err
		}
		for _, it := range items {
			if listLabel != "" && !containsString(it.Labels, listLabel) {
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%.0f%%\t%s\t%s\n", it.Handle.ID, it.Status, it.Progress*100, strings.Join(it.Labels, ","), it.Title)
		}
		return nil
	}}
	list.Flags().StringVar(&listLabel, "label", "", "filter downloads by label")
	dl.AddCommand(list)
	dl.AddCommand(&cobra.Command{Use: "dirs", Short: "List configured download directory aliases", RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := e.load("download")
		if err != nil {
			return err
		}
		keys := make([]string, 0, len(cfg.Download.DirAliases))
		for key := range cfg.Download.DirAliases {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", key, cfg.Download.DirAliases[key])
		}
		return nil
	}})
	var url string
	var dir string
	add := &cobra.Command{Use: "add [URL]", Short: "Add magnet/torrent URL", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := e.load("download")
		if err != nil {
			return err
		}
		if url == "" && len(args) > 0 {
			url = args[0]
		}
		resolvedDir := dir
		if aliasDir, ok := cfg.Download.DirAliases[dir]; ok {
			resolvedDir = aliasDir
		}
		d, err := e.newDownloader(cfg)
		if err != nil {
			return err
		}
		labels := defaultDownloadLabels(cfg)
		h, err := d.Add(cmd.Context(), download.AddDownloadRequest{URL: url, DownloadDir: resolvedDir, Labels: labels})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added %s:%s\n", h.Provider, h.ID)
		return nil
	}}
	add.Flags().StringVar(&url, "url", "", "magnet or torrent URL")
	add.Flags().StringVar(&dir, "dir", "", "download directory override or alias")
	dl.AddCommand(add)
	return dl
}

var titleYearRE = regexp.MustCompile(`^(.+?)\s*\(?([12][0-9]{3})\)?$`)

func splitTitleYear(q string) (string, int) {
	q = strings.TrimSpace(q)
	m := titleYearRE.FindStringSubmatch(q)
	if len(m) != 3 {
		return q, 0
	}
	y, _ := strconv.Atoi(m[2])
	return strings.TrimSpace(m[1]), y
}

func defaultDownloadLabels(cfg *config.Config) []string {
	if cfg.Download.Label == nil || *cfg.Download.Label == "" {
		return nil
	}
	return []string{*cfg.Download.Label}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func (e *commandEnv) searchCommand() *cobra.Command {
	return &cobra.Command{Use: "search QUERY", Short: "Search configured movie release providers", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
		defer cancel()
		cfg, err := e.load("search")
		if err != nil {
			return err
		}
		title, year := splitTitleYear(strings.Join(args, " "))
		result := e.searchReleases(ctx, cfg, search.MovieSearchQuery{Title: title, Year: year})
		for _, err := range result.Errors {
			fmt.Fprintf(cmd.OutOrStdout(), "WARN provider error: provider=%s stage=%s message=%s\n", err.Provider, err.Stage, err.Message)
		}

		for i, r := range result.Releases {
			fmt.Fprintf(cmd.OutOrStdout(),
				"%d. %s (%s, seeders=%d, %.2fGB) %s\n",
				i+1,
				r.Title,
				r.Provider,
				r.Seeders,
				// TODO: size in human-readable format, not fixed GB
				float64(r.SizeBytes)/(1024*1024*1024),
				r.URL)
		}
		return nil
	}}
}

func (e *commandEnv) watchedCommand() *cobra.Command {
	var format string
	cmd := &cobra.Command{Use: "watched", Short: "List watched movies", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := e.load("watched")
		if err != nil {
			return err
		}
		provider, err := e.newWatchedProvider(cfg)
		if err != nil {
			return err
		}
		items, err := provider.ListWatchedMovies(cmd.Context())
		if err != nil {
			return err
		}
		switch format {
		case "", "json":
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(items)
		case "text":
			for _, item := range items {
				parts := []string{item.Title}
				if item.Year > 0 {
					parts = append(parts, strconv.Itoa(item.Year))
				}
				if item.Rating != nil {
					parts = append(parts, fmt.Sprintf("%.1f", *item.Rating))
				}
				if item.WatchedAt != nil {
					parts = append(parts, item.WatchedAt.UTC().Format(time.RFC3339))
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.Join(parts, "\t"))
			}
			return nil
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	}}
	cmd.Flags().StringVar(&format, "format", "json", "output format: json or text")
	return cmd
}

func (e *commandEnv) definitionsCommand() *cobra.Command {
	defs := &cobra.Command{Use: "definitions", Short: "Indexer definition commands"}
	defs.AddCommand(&cobra.Command{Use: "sync", Short: "Download public indexer definitions", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		sync := e.definitionsSync
		if sync == nil {
			sync = tomagnetlib.SyncDefinitions
		}
		m, err := sync()
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "synced %d definitions\n", len(m.Files))
		return nil
	}})
	return defs
}

func (e *commandEnv) newDownloader(cfg *config.Config) (download.Downloader, error) {
	if e.downloader != nil {
		return e.downloader(cfg)
	}
	return newDownloader(cfg)
}

func (e *commandEnv) newSearchProviders(cfg *config.Config) ([]search.Provider, []search.ProviderError) {
	if e.searchProviders != nil {
		return e.searchProviders(cfg)
	}
	return newSearchProviders(cfg)
}

func (e *commandEnv) newWatchedProvider(cfg *config.Config) (watched.Provider, error) {
	if e.watchedProvider != nil {
		return e.watchedProvider(cfg)
	}
	return newWatchedProvider(cfg)
}

func (e *commandEnv) searchReleases(ctx context.Context, cfg *config.Config, q search.MovieSearchQuery) search.Result {
	providers, errs := e.newSearchProviders(cfg)
	result := search.Result{Errors: errs}
	for i, p := range providers {
		rels, err := p.SearchMovie(ctx, q)
		if err != nil {
			result.Errors = append(result.Errors, search.ProviderError{Provider: providerName(i, p, rels), Stage: "search", Message: err.Error()})
			continue
		}
		result.Releases = append(result.Releases, rels...)
	}
	return result
}

func providerName(i int, p search.Provider, rels []search.Release) string {
	if named, ok := p.(interface{ Name() string }); ok && named.Name() != "" {
		return named.Name()
	}
	for _, r := range rels {
		if r.Provider != "" {
			return r.Provider
		}
	}
	return fmt.Sprintf("provider-%d", i+1)
}
