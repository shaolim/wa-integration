package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("hello world")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// verify the webhook
	r.Get("/webhook", func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")

		if mode == "subscribe" && token == verifyToken {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, challenge)
			return
		}

		http.Error(w, "Verification failed", http.StatusForbidden)
	})

	// handle the webhook
	r.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("go")

		// Read the raw body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Failed to read body")
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		fmt.Println("go here")

		// Validate the HMAC signature
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			fmt.Println("webhook POST: missing X-Hub-Signature-256 header")
			log.Println("webhook POST: missing X-Hub-Signature-256 header")
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}
		fmt.Println("go here 1")

		if os.Getenv("WHATSAPP_APP_SECRET") == "" {
			fmt.Println("webhook POST: WHATSAPP_APP_SECRET is not set")
			log.Println("webhook POST: WHATSAPP_APP_SECRET is not set")
			http.Error(w, "Server misconfiguration", http.StatusInternalServerError)
			return
		}

		fmt.Println("go here 2")
		if !validateSignature(body, signature) {
			fmt.Printf("webhook POST: signature mismatch (got %s) \n", signature)
			log.Printf("webhook POST: signature mismatch (got %s)", signature)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		fmt.Println("success")

		// Respond 200 immediately (process async in production)
		w.WriteHeader(http.StatusOK)

		// go processPayload(body)
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}

// Validate HMAC-SHA256 signature
func validateSignature(body []byte, signature string) bool {
	appSecret := os.Getenv("WHATSAPP_APP_SECRET")
	if appSecret == "" || signature == "" {
		return false
	}

	// Remove "sha256=" prefix
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}
	expectedHash := signature[7:]

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	computedHash := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(computedHash), []byte(expectedHash))
}

func processPayload(body []byte) {
	// Parse and handle the webhook payload
	fmt.Printf("Received webhook: %s\n", string(body))

	// Example: unmarshal into a struct
	// var payload WebhookPayload
	// json.Unmarshal(body, &payload)
	// handle messages, statuses, etc.
}
