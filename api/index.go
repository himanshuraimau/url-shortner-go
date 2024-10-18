package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type URL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	CreatedAt   string `json:"created_at"`
}

var urlDB = make(map[string]URL)
var mutex = &sync.Mutex{}

func generateShortURL(OriginalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(OriginalURL))
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	return hash[:6]
}

func short_url(OriginalURL string) string {
	mutex.Lock()
	defer mutex.Unlock()
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
	mutex.Lock()
	defer mutex.Unlock()
	url, ok := urlDB[id]
	if !ok {
		return URL{}, fmt.Errorf("URL not found")
	}
	return url, nil
}

// Handler is the main entry point for the serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var request struct {
			URL string `json:"url"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shortURL := short_url(request.URL)
		response := struct {
			ShortURL string `json:"short_url"`
		}{ShortURL: shortURL}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "GET":
		id := strings.TrimPrefix(r.URL.Path, "/api/")
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
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}