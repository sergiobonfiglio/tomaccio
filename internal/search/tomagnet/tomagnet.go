package tomagnet

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sergiobonfiglio/tomaccio/internal/definitions"
	"github.com/sergiobonfiglio/tomaccio/internal/search"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

type Client struct {
	name           string
	indexerID      string
	baseURL        string
	timeoutSeconds int
	settings       map[string]string
}

func New(name, indexerID, baseURL string, timeoutSeconds int, settings map[string]string) *Client {
	return &Client{
		name:           name,
		indexerID:      indexerID,
		baseURL:        baseURL,
		timeoutSeconds: timeoutSeconds,
		settings:       maps.Clone(settings),
	}
}

func (c *Client) Name() string { return c.name }

func (c *Client) SearchMovie(ctx context.Context, q search.MovieSearchQuery) ([]search.Release, error) {
	definition, err := definitions.LoadByID(c.indexerID)
	if err != nil {
		return nil, err
	}
	if err := applyDefinitionSettings(definition, c.settings); err != nil {
		return nil, err
	}

	resp := tomagnetlib.Search(ctx, tomagnetlib.SearchOptions{
		Query: tomagnetlib.Query{Movie: &tomagnetlib.MovieQuery{
			Title:  q.Title,
			Year:   q.Year,
			IMDBID: q.IMDBID,
			TMDBID: searchID(q.TMDBID),
		}},
		Indexers: []tomagnetlib.Indexer{{
			ID:             c.indexerID,
			BaseURL:        c.baseURL,
			TimeoutSeconds: c.timeoutSeconds,
			Definition:     definition,
		}},
	})
	if len(resp.Errors) > 0 {
		err := resp.Errors[0]
		message := redactSettingValues(err.Message, c.settings)
		return nil, fmt.Errorf("tomagnet %s %s: %s", c.name, err.Stage, message)
	}
	out := make([]search.Release, 0, len(resp.Results))
	for _, result := range resp.Results {
		url := firstNonEmpty(result.MagnetURL, result.DownloadURL, result.DetailsURL)
		if url == "" || result.Title == "" {
			continue
		}
		release := search.Release{
			Provider:  c.name,
			Title:     result.Title,
			URL:       url,
			SizeBytes: result.SizeBytes,
			Seeders:   result.Seeders,
			Leechers:  result.Leechers,
			Category:  result.Category,
		}
		if result.PublishedAt != nil {
			release.Published = result.PublishedAt.UTC().Format(time.RFC3339)
		}
		if result.EnrichmentError != nil {
			release.ResolutionError = redactSettingValues(result.EnrichmentError.Message, c.settings)
		}
		out = append(out, release)
	}
	return out, nil
}

func applyDefinitionSettings(definition *tomagnetlib.Definition, values map[string]string) error {
	known := make(map[string]bool, len(definition.Settings))
	for i := range definition.Settings {
		name := definition.Settings[i].Name
		known[name] = true
		if value, ok := values[name]; ok {
			definition.Settings[i].Default = value
		}
	}
	for name := range values {
		if !known[name] {
			return fmt.Errorf("indexer %q does not declare setting %q", definition.ID, name)
		}
	}
	return nil
}

func redactSettingValues(message string, settings map[string]string) string {
	encoded := map[string]bool{}
	for _, value := range settings {
		if value == "" {
			continue
		}
		encoded[value] = true
		encoded[url.QueryEscape(value)] = true
		encoded[url.PathEscape(value)] = true
	}
	values := make([]string, 0, len(encoded))
	for value := range encoded {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	for _, value := range values {
		message = strings.ReplaceAll(message, value, "[REDACTED]")
	}
	return message
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func searchID(id int) string {
	if id <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", id)
}
