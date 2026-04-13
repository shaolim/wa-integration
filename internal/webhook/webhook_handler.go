package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/shaolim/wa-integration/pkg/whatsapp"
)

// TODO: create WA client

func (w *Webhook) VerifyWebhook(rw http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == w.conf.WAVerifyToken {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, challenge)
		return
	}

	http.Error(rw, "Verification failed", http.StatusForbidden)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body.Write(b)
	return rr.ResponseWriter.Write(b)
}

func (w *Webhook) HandleWebhook(rw http.ResponseWriter, r *http.Request) {
	rr := &responseRecorder{ResponseWriter: rw, statusCode: http.StatusOK}
	defer func() {
		slog.Info("webhook POST", slog.Int("status", rr.statusCode), slog.String("body", rr.body.String()))
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rr, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	slog.Info("webhook POST: received", slog.String("body", string(body)))

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		slog.Warn("webhook POST: missing X-Hub-Signature-256 header")
		http.Error(rr, "Missing signature", http.StatusUnauthorized)
		return
	}

	if !w.waClient.ValidateSignature(body, signature) {
		slog.Warn("webhook POST: signature mismatch", slog.String("signature", signature))
		http.Error(rr, "Invalid signature", http.StatusUnauthorized)
		return
	}

	rr.WriteHeader(http.StatusOK)
	fmt.Fprint(rr, "OK")

	go w.processPayload(body)
}

func (w *Webhook) processPayload(body []byte) {
	payload, err := whatsapp.Parse(body)
	if err != nil {
		slog.Error("processPayload: failed to parse body", slog.Any("error", err))
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				mediaID := whatsapp.MediaIDFromMessage(msg)
				if mediaID == "" {
					continue
				}

				data, mimeType, err := w.waClient.DownloadMedia(mediaID)
				if err != nil {
					slog.Error("processPayload: download media", slog.String("media_id", mediaID), slog.Any("error", err))
					continue
				}

				key := "media/" + mediaID
				if err := w.uploader.Upload(context.Background(), key, data, mimeType); err != nil {
					slog.Error("processPayload: upload media", slog.String("media_id", mediaID), slog.Any("error", err))
					continue
				}

				slog.Info("processPayload: uploaded media",
					slog.String("media_id", mediaID),
					slog.String("s3_key", key),
					slog.String("mime_type", mimeType),
					slog.Int("size", len(data)),
				)
			}
		}
	}
}
