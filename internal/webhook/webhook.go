package webhook

import (
	"context"

	"github.com/shaolim/wa-integration/internal/config"
	"github.com/shaolim/wa-integration/pkg/whatsapp"
)

// Uploader stores a media file at the given key.
type Uploader interface {
	Upload(ctx context.Context, key string, data []byte, mimeType string) error
}

type Webhook struct {
	conf     *config.Config
	uploader Uploader
	waClient *whatsapp.WAClient
}

func New(conf *config.Config, uploader Uploader, waClient *whatsapp.WAClient) *Webhook {
	return &Webhook{
		conf:     conf,
		uploader: uploader,
		waClient: waClient,
	}
}
