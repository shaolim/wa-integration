package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaolim/wa-integration/internal/config"
	"github.com/shaolim/wa-integration/internal/storage"
	"github.com/shaolim/wa-integration/internal/webhook"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		panic(err)
	}

	uploader, err := storage.NewS3Uploader(
		context.Background(),
		cfg.S3Region,
		cfg.S3BucketName,
		cfg.S3AccessKey,
		cfg.S3SecretAccessKey,
	)
	if err != nil {
		slog.Error("failed to create s3 uploader", slog.Any("error", err))
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	wh := webhook.New(cfg, uploader)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/webhook", wh.VerifyWebhook)
	r.Post("/webhook", wh.HandleWebhook)

	err = http.ListenAndServe(":8080", r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}
