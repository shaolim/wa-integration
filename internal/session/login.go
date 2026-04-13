package session

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/shaolim/wa-integration/web"
)

var loginTmpl = template.Must(template.New("login").Parse(web.LoginHTML))

const sessionCookieName = "session_token"
const sessionTTL = 24 * time.Hour

type Handler struct {
	username string
	password string

	mu       sync.Mutex
	sessions map[string]time.Time
}

func NewHandler(username, password string) *Handler {
	return &Handler{
		username: username,
		password: password,
		sessions: make(map[string]time.Time),
	}
}

func (h *Handler) ServeLogin(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	loginTmpl.Execute(rw, nil)
}

func (h *Handler) HandleLogin(rw http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != h.username || password != h.password {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusUnauthorized)
		loginTmpl.Execute(rw, map[string]string{"Error": "Invalid username or password"})
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.sessions[token] = time.Now().Add(sessionTTL)
	h.mu.Unlock()

	http.SetCookie(rw, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(sessionTTL),
	})

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

// RequireAuth is a middleware that redirects unauthenticated requests to /login.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		h.mu.Lock()
		expiry, ok := h.sessions[cookie.Value]
		h.mu.Unlock()

		if !ok || time.Now().After(expiry) {
			h.mu.Lock()
			delete(h.sessions, cookie.Value)
			h.mu.Unlock()
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
