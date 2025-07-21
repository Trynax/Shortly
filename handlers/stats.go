package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/trynax/shortly/storage"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the short code from URL path
	// Expected format: /stats/{shortCode}
	path := strings.TrimPrefix(r.URL.Path, "/stats/")
	if path == "" || path == "stats" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	store := storage.GetStore()
	if store == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	urlStats, err := store.GetURLStats(path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "URL not found or expired", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get URL stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urlStats)
}
