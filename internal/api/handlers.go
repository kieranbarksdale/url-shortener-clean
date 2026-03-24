package api

import (
	"fmt"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"url_shortener/internal/shortener" 
	"url_shortener/internal/cache"
	"github.com/redis/go-redis/v9"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code string `json:"code"`
}

func ShortenHandler(w http.ResponseWriter, req *http.Request, db *sql.DB, redisClient *redis.Client) { 

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if limited, err := cache.IsRateLimited(redisClient, ip); err != nil {
		http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
		return
	} else if limited {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	var request ShortenRequest
	if req.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil || request.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	} 
	
	if !strings.Contains(request.URL, "https://") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	code, err := shortener.GenerateCode(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = shortener.StoreURL(db, code, request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ShortenResponse{Code: code})
}


func RedirectHandler(w http.ResponseWriter, req *http.Request, db *sql.DB, redisClient *redis.Client) {
	if req.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	code := req.URL.Path[1:]  

	redirectURL, err := cache.GetURL(redisClient, code)
	if err != nil {
		redirectURL, err = shortener.GetURL(db, code)
		if err == sql.ErrNoRows {
			http.Error(w, "Short code not found", http.StatusNotFound)
			return
		} 
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = cache.SetURL(redisClient, code, redirectURL)
		if err != nil {
			fmt.Println("Error setting URL in cache:", err)
		}
	}


	http.Redirect(w, req, redirectURL, http.StatusFound)
	
}

