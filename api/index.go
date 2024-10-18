package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type URL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	CreatedAt   string `json:"created_at"`
}

var urlDB = make(map[string]URL)

func generateShortURL(OriginalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(OriginalURL))
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	return hash[:6]
}

func short_url(OriginalURL string) string {
	for {
		shortURL := generateShortURL(OriginalURL)
		id := shortURL
		if _, exists := urlDB[id]; !exists {
			urlDB[id] = URL{
				ID:          id,
				OriginalURL: OriginalURL,
				ShortURL:    shortURL,
				CreatedAt:   time.Now().String(),
			}
			return shortURL
		}
	}
}

func getURL(id string) (URL, error) {
	url, ok := urlDB[id]
	if !ok {
		return URL{}, fmt.Errorf("URL not found")
	}
	return url, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received")
}

func shortUrlHandler(w http.ResponseWriter, r *http.Request) {
	var response struct {
		URL string `json:"short_url"`
	}
	err := json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := short_url(response.URL)

	data := struct {
		URL string `json:"short_url"`
	}{URL: shortURL}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/redirect/")
	if id == "" {
		http.Error(w, "Missing short URL ID", http.StatusBadRequest)
		return
	}

	url, err := getURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/shorten", shortUrlHandler)
	http.HandleFunc("/redirect/", redirectHandler)

	fmt.Println("Starting the server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}
