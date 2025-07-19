package handlers


import (
	 "net/http"
	 "strings"
	 "github.com/trynax/shortly/storage"
)

func RedirectHandler(w http.ResponseWriter, r  *http.Request){
	code := strings.TrimPrefix(r.URL.Path, "/")
	if longURL,exists := storage.URLSTORE[code]; exists {
		http.Redirect(w,r, longURL, http.StatusFound)
		return
	}

	http.NotFound(w,r)
}