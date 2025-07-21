package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trynax/shortly/handlers"
	"github.com/trynax/shortly/storage"
)

func main() {
	// Initialize database
	store, err := storage.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	// Start periodic cleanup of expired URLs (every hour)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := store.CleanupExpiredURLs(); err != nil {
					log.Printf("Error during periodic cleanup: %v", err)
				}
			}
		}
	}()

	http.HandleFunc("/shorten", handlers.ShortenHandler)
	http.HandleFunc("/stats/", handlers.StatsHandler)
	http.HandleFunc("/", handlers.RedirectHandler)

	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Database initialized successfully!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
