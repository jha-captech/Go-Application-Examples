package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define server port
	port := "8080"

	// Register handler
	http.HandleFunc("GET /hello-world", helloHandler)

	// Start the server
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// HelloHandler responds with a hello world message
func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Hello, World!"}`)
}
