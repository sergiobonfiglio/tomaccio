package download

import "context"

// Downloader is the provider-agnostic contract used by orchestration code.
type Downloader interface {
	Test(ctx context.Context) error
	Add(ctx context.Context, req AddDownloadRequest) (*DownloadHandle, error)
	List(ctx context.Context) ([]DownloadStatus, error)
	Get(ctx context.Context, handle DownloadHandle) (*DownloadStatus, error)
	Cancel(ctx context.Context, handle DownloadHandle) error
}

type AddDownloadRequest struct {
	URL         string   `json:"url"`
	DownloadDir string   `json:"download_dir"`
	Labels      []string `json:"labels"`
	Title       string   `json:"title"`
}

type DownloadHandle struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
}

type DownloadStatus struct {
	Handle   DownloadHandle `json:"handle"`
	Title    string         `json:"title"`
	Status   string         `json:"status"`
	Progress float64        `json:"progress"`
	Labels   []string       `json:"labels"`
}
