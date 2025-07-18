package models



type URL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	Clicks   int    `json:"clicks"`
	CreatedAt string `json:"created_at"`
}