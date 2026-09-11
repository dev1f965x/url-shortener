// Command url-shortener serves a small URL shortening API backed by SQLite.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

const (
	codeLength  = 7
	codeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// codeAttempts bounds retries when a random code collides with an existing one.
	codeAttempts = 5
	maxBodyBytes = 1 << 16
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type server struct {
	db *sql.DB
	// baseURL prefixes returned links. Empty means the request's own scheme and host.
	baseURL string
}

func main() {
	addr := env("ADDR", ":8080")
	dbPath := env("DB_PATH", "data/urls.db")

	db, err := openDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := &server{db: db, baseURL: os.Getenv("BASE_URL")}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", s.handleShorten)
	mux.HandleFunc("GET /{code}", s.handleRedirect)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Print(err)
	}
}

// openDB opens the SQLite file at path, creating it and its schema if needed.
func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite allows one writer at a time; a single connection avoids SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_url TEXT NOT NULL UNIQUE,
			short_code TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (s *server) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if !isValidURL(req.URL) {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}

	code, created, err := s.shorten(r.Context(), req.URL)
	if err != nil {
		log.Printf("shorten %q: %v", req.URL, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, shortenResponse{
		ShortCode:   code,
		ShortURL:    s.shortURL(r, code),
		OriginalURL: req.URL,
	})
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	var originalURL string
	err := s.db.QueryRowContext(r.Context(),
		"SELECT original_url FROM urls WHERE short_code = ?", code,
	).Scan(&originalURL)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.NotFound(w, r)
	case err != nil:
		log.Printf("redirect %q: %v", code, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

// shorten returns the code for originalURL, storing a new one if the URL is unseen.
// created reports whether a new row was inserted.
func (s *server) shorten(ctx context.Context, originalURL string) (code string, created bool, err error) {
	for range codeAttempts {
		err = s.db.QueryRowContext(ctx,
			"SELECT short_code FROM urls WHERE original_url = ?", originalURL,
		).Scan(&code)
		if err == nil {
			return code, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", false, err
		}

		code = randomCode()
		res, insertErr := s.db.ExecContext(ctx,
			`INSERT INTO urls (original_url, short_code) VALUES (?, ?)
			 ON CONFLICT (original_url) DO NOTHING`,
			originalURL, code,
		)
		if insertErr != nil {
			// Most likely a short_code collision; try another code.
			continue
		}
		if n, _ := res.RowsAffected(); n == 1 {
			return code, true, nil
		}
		// A concurrent request stored the same URL first; the next pass reads its code.
	}
	return "", false, errors.New("could not allocate a unique short code")
}

// shortURL builds the public link for code.
func (s *server) shortURL(r *http.Request, code string) string {
	base := s.baseURL
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	return strings.TrimSuffix(base, "/") + "/" + code
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func randomCode() string {
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = codeCharset[rand.IntN(len(codeCharset))]
	}
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
