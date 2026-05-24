package watched

import (
	"context"
	"time"
)

// Provider lists watched media from an external library.
type Provider interface {
	ListWatchedMovies(ctx context.Context) ([]Item, error)
}

// Item is a watched movie entry.
type Item struct {
	Title     string     `json:"title"`
	Year      int        `json:"year,omitempty"`
	Rating    *float64   `json:"rating,omitempty"`
	WatchedAt *time.Time `json:"watched_at,omitempty"`
}
