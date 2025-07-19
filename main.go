package main

import (
	"fmt"
	"net/http"
	"github.com/trynax/shortly/handlers"
)


func main (){

	http.HandleFunc("/shorten", handlers.ShortenHandler)
	http.HandleFunc("/", handlers.RedirectHandler)


	fmt.Println("server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}