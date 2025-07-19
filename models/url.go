package models



type URL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
	Clicks   int    `json:"clicks"`
	CreatedAt string `json:"created_at"`
}

type RequestBody struct {
	URL string `json:"url`
}


type ResponseBody struct {
	ShortCode string `json: "short_code"`
	ShortURL string `json:short_url`
}