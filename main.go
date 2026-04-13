package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaolim/wa-integration/internal/config"
	"github.com/shaolim/wa-integration/internal/webhook"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	webhook := webhook.New(config)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/webhook", webhook.VerifyWebhook)
	r.Post("/webhook", webhook.HandleWebhook)

	err = http.ListenAndServe(":8080", r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}
