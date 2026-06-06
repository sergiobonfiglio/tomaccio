package tomagnet

import (
	"context"
	"fmt"
	"time"

	"github.com/sergiobonfiglio/tomaccio/internal/search"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

type Client struct {
	name           string
	indexerID      string
	baseURL        string
	timeoutSeconds int
}

func New(name, indexerID, baseURL string, timeoutSeconds int) *Client {
	return &Client{name: name, indexerID: indexerID, baseURL: baseURL, timeoutSeconds: timeoutSeconds}
}

func (c *Client) Name() string { return c.name }

func (c *Client) SearchMovie(ctx context.Context, q search.MovieSearchQuery) ([]search.Release, error) {
	definition, err := tomagnetlib.LoadDefinitionByID(c.indexerID)
	if err != nil {
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
		return nil, fmt.Errorf("tomagnet %s %s: %s", c.name, err.Stage, err.Message)
	}
	out := make([]search.Release, 0, len(resp.Results))
	for _, result := range resp.Results {
		url := firstNonEmpty(result.MagnetURL, result.DownloadURL)
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
		out = append(out, release)
	}
	return out, nil
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
