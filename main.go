package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/api/translate", handleTranslate)
	http.HandleFunc("/api/dictionary", handleDictionary)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Server is listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
