package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sergiobonfiglio/tomaccio/internal/download"
)

const sessionHeader = "X-Transmission-Session-Id"

type Client struct {
	url         string
	username    string
	password    string
	downloadDir string
	httpClient  *http.Client
	mu          sync.Mutex
	sessionID   string
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

func New(url, username, password, downloadDir string, opts ...Option) *Client {
	c := &Client{url: url, username: username, password: password, downloadDir: downloadDir, httpClient: &http.Client{Timeout: 30 * time.Second}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Test(ctx context.Context) error {
	_, err := c.rpc(ctx, "session-get", nil)
	return err
}

func (c *Client) Add(ctx context.Context, req download.AddDownloadRequest) (*download.DownloadHandle, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, errors.New("download URL is required")
	}
	args := map[string]any{"filename": req.URL}
	if dir := firstNonEmpty(req.DownloadDir, c.downloadDir); dir != "" {
		args["download-dir"] = dir
	}
	body, err := c.rpc(ctx, "torrent-add", args)
	if err != nil {
		return nil, err
	}
	var out struct {
		Arguments struct {
			TorrentAdded struct {
				ID         int    `json:"id"`
				HashString string `json:"hashString"`
			} `json:"torrent-added"`
			TorrentDuplicate struct {
				ID         int    `json:"id"`
				HashString string `json:"hashString"`
			} `json:"torrent-duplicate"`
		} `json:"arguments"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	id := out.Arguments.TorrentAdded.ID
	if id == 0 {
		id = out.Arguments.TorrentDuplicate.ID
	}
	if id == 0 {
		return nil, fmt.Errorf("transmission did not return torrent id")
	}
	return &download.DownloadHandle{Provider: "transmission", ID: strconv.Itoa(id)}, nil
}

func (c *Client) List(ctx context.Context) ([]download.DownloadStatus, error) {
	statuses, err := c.torrentGet(ctx, []string{"id", "name", "percentDone", "status"}, nil)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Client) Get(ctx context.Context, handle download.DownloadHandle) (*download.DownloadStatus, error) {
	id, err := strconv.Atoi(handle.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid transmission id %q", handle.ID)
	}
	statuses, err := c.torrentGet(ctx, []string{"id", "name", "percentDone", "status"}, []int{id})
	if err != nil {
		return nil, err
	}
	if len(statuses) == 0 {
		return nil, fmt.Errorf("download %s not found", handle.ID)
	}
	return &statuses[0], nil
}

func (c *Client) Cancel(ctx context.Context, handle download.DownloadHandle) error {
	id, err := strconv.Atoi(handle.ID)
	if err != nil {
		return fmt.Errorf("invalid transmission id %q", handle.ID)
	}
	_, err = c.rpc(ctx, "torrent-remove", map[string]any{"ids": []int{id}, "delete-local-data": false})
	return err
}

func (c *Client) torrentGet(ctx context.Context, fields []string, ids []int) ([]download.DownloadStatus, error) {
	args := map[string]any{"fields": fields}
	if len(ids) > 0 {
		args["ids"] = ids
	}
	body, err := c.rpc(ctx, "torrent-get", args)
	if err != nil {
		return nil, err
	}
	var out struct {
		Arguments struct {
			Torrents []struct {
				ID          int     `json:"id"`
				Name        string  `json:"name"`
				PercentDone float64 `json:"percentDone"`
				Status      int     `json:"status"`
			} `json:"torrents"`
		} `json:"arguments"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	statuses := make([]download.DownloadStatus, 0, len(out.Arguments.Torrents))
	for _, t := range out.Arguments.Torrents {
		statuses = append(statuses, download.DownloadStatus{Handle: download.DownloadHandle{Provider: "transmission", ID: strconv.Itoa(t.ID)}, Title: t.Name, Progress: t.PercentDone, Status: mapStatus(t.Status, t.PercentDone)})
	}
	return statuses, nil
}

func (c *Client) rpc(ctx context.Context, method string, args any) ([]byte, error) {
	payload, err := json.Marshal(map[string]any{"method": method, "arguments": args})
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.username != "" || c.password != "" {
			req.SetBasicAuth(c.username, c.password)
		}
		c.mu.Lock()
		sid := c.sessionID
		c.mu.Unlock()
		if sid != "" {
			req.Header.Set(sessionHeader, sid)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if resp.StatusCode == http.StatusConflict && resp.Header.Get(sessionHeader) != "" {
			// Transmission intentionally returns 409 until callers echo this token.
			c.mu.Lock()
			c.sessionID = resp.Header.Get(sessionHeader)
			c.mu.Unlock()
			continue
		}
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("transmission rpc status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		var envelope struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, err
		}
		if envelope.Result != "success" {
			return nil, fmt.Errorf("transmission rpc %s failed: %s", method, envelope.Result)
		}
		return body, nil
	}
	return nil, errors.New("transmission session negotiation failed")
}

func mapStatus(status int, progress float64) string {
	if progress >= 1 {
		return "completed"
	}
	switch status {
	case 0:
		return "stopped"
	case 4:
		return "downloading"
	case 6:
		return "seeding"
	default:
		return "active"
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
