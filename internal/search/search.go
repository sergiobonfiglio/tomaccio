package search

import "context"

type Provider interface {
	SearchMovie(ctx context.Context, query MovieSearchQuery) ([]Release, error)
}

type ProviderError struct {
	Provider string `json:"provider"`
	Stage    string `json:"stage"`
	Message  string `json:"message"`
}

type Result struct {
	Releases []Release       `json:"releases"`
	Errors   []ProviderError `json:"errors"`
}

type MovieSearchQuery struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	IMDBID string `json:"imdb_id"`
	TMDBID int    `json:"tmdb_id"`
}

type Release struct {
	Provider  string `json:"provider"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	SizeBytes int64  `json:"size_bytes"`
	Seeders   int    `json:"seeders"`
	Leechers  int    `json:"leechers"`
	Category  string `json:"category"`
	Published string `json:"published"`
}
