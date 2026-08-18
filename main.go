package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	_ "modernc.org/sqlite"
)

const shortCodeLength = 7
const shortCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var db *sql.DB

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "/app/data/urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initSchema(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/", handleRedirect)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_url TEXT NOT NULL UNIQUE,
			short_code TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if !isValidURL(req.URL) {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}

	code, status, err := getOrCreateShortCode(req.URL)
	if err != nil {
		log.Println("shorten error:", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := shortenResponse{
		ShortCode:   code,
		ShortURL:    "http://localhost:8080/" + code,
		OriginalURL: req.URL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = ?", code).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println("redirect error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// getOrCreateShortCode returns an existing code for originalURL if one exists
// (http.StatusOK), otherwise generates a new one and inserts it (http.StatusCreated).
func getOrCreateShortCode(originalURL string) (code string, status int, err error) {
	err = db.QueryRow("SELECT short_code FROM urls WHERE original_url = ?", originalURL).Scan(&code)
	if err == nil {
		return code, http.StatusOK, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", 0, err
	}

	for i := 0; i < 5; i++ {
		code = randomCode(shortCodeLength)
		_, err = db.Exec(
			"INSERT INTO urls (original_url, short_code) VALUES (?, ?)",
			originalURL, code,
		)
		if err == nil {
			return code, http.StatusCreated, nil
		}
	}
	return "", 0, errors.New("failed to generate unique short code after retries")
}

func randomCode(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = shortCodeCharset[r.Intn(len(shortCodeCharset))]
	}
	return string(b)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
