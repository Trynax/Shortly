package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/trynax/shortly/models"
	"github.com/trynax/shortly/storage"
	"github.com/trynax/shortly/utils"
)

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var body models.RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil || strings.TrimSpace(body.URL) == "" {
		http.Error(w, "Invalid JSON or empty URL", http.StatusBadRequest)
		return
	}

	// Generate a unique short code
	code := utils.GenerateCode(6)

	// Get storage instance and save URL
	store := storage.GetStore()
	if store == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	err = store.SaveURL(code, body.URL)
	if err != nil {
		log.Printf("Error saving URL: %v", err)
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	resp := models.ResponseBody{
		ShortCode: code,
		ShortURL:  "http://localhost:8080/" + code,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
