package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/trynax/shortly/models"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

var store *Storage

// Initialize creates and returns a new Storage instance
func Init() (*Storage, error) {
	db, err := sql.Open("sqlite", "./shortly.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	s := &Storage{db: db}

	// Create tables if they don't exist
	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	store = s
	return s, nil
}

// createTables creates the necessary database tables and handles migrations
func (s *Storage) createTables() error {
	// First, create the table without expires_at for compatibility
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT UNIQUE NOT NULL,
		long_url TEXT NOT NULL,
		clicks INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create urls table: %v", err)
	}

	// Check if expires_at column exists
	var columnExists bool
	checkColumnQuery := `PRAGMA table_info(urls);`
	rows, err := s.db.Query(checkColumnQuery)
	if err != nil {
		return fmt.Errorf("failed to check table info: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		if name == "expires_at" {
			columnExists = true
			break
		}
	}

	// Add expires_at column if it doesn't exist
	if !columnExists {
		log.Println("Adding expires_at column to existing urls table...")
		addColumnQuery := `ALTER TABLE urls ADD COLUMN expires_at DATETIME;`
		_, err = s.db.Exec(addColumnQuery)
		if err != nil {
			return fmt.Errorf("failed to add expires_at column: %v", err)
		}

		// Update existing rows to have an expiration time (5 hours from creation)
		updateQuery := `UPDATE urls SET expires_at = datetime(created_at, '+5 hours') WHERE expires_at IS NULL;`
		_, err = s.db.Exec(updateQuery)
		if err != nil {
			return fmt.Errorf("failed to update existing rows with expiration: %v", err)
		}

		log.Println("Migration completed successfully!")
	}

	return nil
}

// SaveURL stores a new URL mapping in the database with 5-hour expiration
func (s *Storage) SaveURL(shortCode, longURL string) error {
	// Set expiration time to 5 hours from now
	expiresAt := time.Now().Add(5 * time.Hour)

	query := `INSERT INTO urls (short_code, long_url, expires_at) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, shortCode, longURL, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save URL: %v", err)
	}
	return nil
}

// GetURL retrieves the long URL by short code if it hasn't expired
func (s *Storage) GetURL(shortCode string) (string, error) {
	var longURL string
	var expiresAt time.Time

	query := `SELECT long_url, expires_at FROM urls WHERE short_code = ?`
	err := s.db.QueryRow(query, shortCode).Scan(&longURL, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("failed to get URL: %v", err)
	}

	// Check if URL has expired
	if time.Now().After(expiresAt) {
		// Clean up expired URL
		s.deleteExpiredURL(shortCode)
		return "", fmt.Errorf("URL has expired")
	}

	return longURL, nil
}

// IncrementClicks increases the click count for a URL
func (s *Storage) IncrementClicks(shortCode string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = ?`
	result, err := s.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment clicks: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

// deleteExpiredURL removes an expired URL from the database
func (s *Storage) deleteExpiredURL(shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = ?`
	_, err := s.db.Exec(query, shortCode)
	if err != nil {
		log.Printf("Error deleting expired URL %s: %v", shortCode, err)
		return err
	}
	log.Printf("Deleted expired URL: %s", shortCode)
	return nil
}

// CleanupExpiredURLs removes all expired URLs from the database
func (s *Storage) CleanupExpiredURLs() error {
	query := `DELETE FROM urls WHERE expires_at < ?`
	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired URLs: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected during cleanup: %v", err)
	}

	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired URLs", rowsAffected)
	}

	return nil
}

// GetURLStats retrieves complete URL information including stats
func (s *Storage) GetURLStats(shortCode string) (*models.URL, error) {
	var url models.URL
	var createdAt time.Time
	var expiresAt time.Time

	query := `SELECT short_code, long_url, clicks, created_at, expires_at FROM urls WHERE short_code = ?`
	err := s.db.QueryRow(query, shortCode).Scan(&url.ShortURL, &url.LongURL, &url.Clicks, &createdAt, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL stats: %v", err)
	}

	// Check if URL has expired
	if time.Now().After(expiresAt) {
		// Clean up expired URL
		s.deleteExpiredURL(shortCode)
		return nil, fmt.Errorf("URL has expired")
	}

	url.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	url.ExpiresAt = expiresAt.Format("2006-01-02 15:04:05")
	return &url, nil
}

// GetAllURLs retrieves all non-expired URLs from the database
func (s *Storage) GetAllURLs() ([]models.URL, error) {
	// First clean up expired URLs
	s.CleanupExpiredURLs()

	query := `SELECT short_code, long_url, clicks, created_at, expires_at FROM urls WHERE expires_at > ? ORDER BY created_at DESC`
	rows, err := s.db.Query(query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get all URLs: %v", err)
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		var createdAt time.Time
		var expiresAt time.Time

		err := rows.Scan(&url.ShortURL, &url.LongURL, &url.Clicks, &createdAt, &expiresAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		url.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		url.ExpiresAt = expiresAt.Format("2006-01-02 15:04:05")
		urls = append(urls, url)
	}

	return urls, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetStore returns the global storage instance
func GetStore() *Storage {
	return store
}
