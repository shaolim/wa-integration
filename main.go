package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/shaolim/wa-integration/internal/config"
	"github.com/shaolim/wa-integration/internal/session"
	"github.com/shaolim/wa-integration/internal/webhook"
	"github.com/shaolim/wa-integration/pkg/storage"
	"github.com/shaolim/wa-integration/pkg/whatsapp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	waClient := whatsapp.New(cfg.WAAccessToken, cfg.WAVerifyToken, cfg.WAAppSecret)

	sess := session.NewHandler(cfg.Username, cfg.UserPassword)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	wh := webhook.New(cfg, uploader, waClient)

	// public routes
	r.Get("/login", sess.ServeLogin)
	r.Post("/login", sess.HandleLogin)
	r.Get("/webhook", wh.VerifyWebhook)
	r.Post("/webhook", wh.HandleWebhook)

	// health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(sess.RequireAuth)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	err = http.ListenAndServe(":8080", r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}
