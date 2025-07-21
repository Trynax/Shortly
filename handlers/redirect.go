package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/trynax/shortly/storage"
)

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	// Skip empty codes or favicon requests
	if code == "" || code == "favicon.ico" {
		http.NotFound(w, r)
		return
	}

	store := storage.GetStore()
	if store == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	longURL, err := store.GetURL(code)
	if err != nil {
		log.Printf("Error getting URL for code %s: %v", code, err)
		http.NotFound(w, r)
		return
	}

	// Increment click counter
	err = store.IncrementClicks(code)
	if err != nil {
		log.Printf("Error incrementing clicks for code %s: %v", code, err)
		// Don't fail the redirect, just log the error
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
