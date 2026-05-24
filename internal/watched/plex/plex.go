package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sergiobonfiglio/tomaccio/internal/watched"
)

// Client reads watched movies from Plex.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithHTTPClient sets HTTP client used by Plex requests.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// New creates a Plex watched provider.
func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, httpClient: &http.Client{Timeout: 30 * time.Second}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ListWatchedMovies returns movies marked watched in all Plex movie libraries.
func (c *Client) ListWatchedMovies(ctx context.Context) ([]watched.Item, error) {
	sections, err := c.movieSections(ctx)
	if err != nil {
		return nil, err
	}
	var out []watched.Item
	for _, section := range sections {
		items, err := c.sectionWatchedMovies(ctx, section.Key)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

type sectionsResponse struct {
	Directories []struct {
		Key  string `xml:"key,attr"`
		Type string `xml:"type,attr"`
	} `xml:"Directory"`
}

func (c *Client) movieSections(ctx context.Context) ([]struct{ Key string }, error) {
	var resp sectionsResponse
	if err := c.getXML(ctx, "/library/sections", nil, &resp); err != nil {
		return nil, err
	}
	sections := make([]struct{ Key string }, 0, len(resp.Directories))
	for _, d := range resp.Directories {
		if d.Type == "movie" && d.Key != "" {
			sections = append(sections, struct{ Key string }{Key: d.Key})
		}
	}
	return sections, nil
}

type moviesResponse struct {
	Videos []struct {
		Title        string  `xml:"title,attr"`
		Year         int     `xml:"year,attr"`
		UserRating   float64 `xml:"userRating,attr"`
		LastViewedAt int64   `xml:"lastViewedAt,attr"`
		ViewCount    int     `xml:"viewCount,attr"`
	} `xml:"Video"`
}

func (c *Client) sectionWatchedMovies(ctx context.Context, key string) ([]watched.Item, error) {
	var resp moviesResponse
	if err := c.getXML(ctx, "/library/sections/"+url.PathEscape(key)+"/all", url.Values{"viewed": {"1"}}, &resp); err != nil {
		return nil, err
	}
	items := make([]watched.Item, 0, len(resp.Videos))
	for _, v := range resp.Videos {
		title := strings.TrimSpace(v.Title)
		if title == "" {
			continue
		}
		item := watched.Item{Title: title, Year: v.Year}
		if v.UserRating > 0 {
			rating := v.UserRating
			item.Rating = &rating
		}
		if v.LastViewedAt > 0 {
			watchedAt := time.Unix(v.LastViewedAt, 0).UTC()
			item.WatchedAt = &watchedAt
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *Client) getXML(ctx context.Context, path string, q url.Values, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("plex url is required")
	}
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return err
	}
	if q == nil {
		q = url.Values{}
	}
	if c.token != "" {
		q.Set("X-Plex-Token", c.token)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("X-Plex-Token", c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if readErr != nil {
		return readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("plex status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := xml.Unmarshal(body, out); err != nil {
		return fmt.Errorf("parse plex xml: %w", err)
	}
	return nil
}
