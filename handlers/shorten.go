package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/trynax/shortly/storage"
	"github.com/trynax/shortly/models"
	"github.com/trynax/shortly/utils"
)


func ShortURL (longUrl string ) (string, error) {
	if strings.Trim(longUrl, " ") == "" {
		return "", errors.New("Can't have nil value as url")
	}
	code := utils.GenerateCode(6)
	storage.URLSTORE[code]= longUrl
	return code, nil

}

func ShortenHandler (w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Only Post allowed", http.StatusMethodNotAllowed)
		return
	}

	var body models.RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil || strings.TrimSpace(body.URL)==""{
		http.Error(w, "Invalid JSON or empty URL", http.StatusBadRequest)
		return
	}

	code := utils.GenerateCode(6)
	storage.URLSTORE[code]=body.URL

	resp := models.ResponseBody{
		ShortCode: code,
		ShortURL: "http://localhost:8080/"+ code,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)


}