package main

import (
	"log"
	"net/http"
)

func main() {
	//Define routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/submit", submitContent)

	//Start server on port 8080
	log.Println("Server starting on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
