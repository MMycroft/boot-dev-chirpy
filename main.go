// Package main holds the main app logic
package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	_ROOT = "."
	_PORT = 8080
)

func main() {
	mux := http.NewServeMux()

	fileServerHandler := http.FileServer(http.Dir(_ROOT))

	mux.Handle("/app/", http.StripPrefix("/app", fileServerHandler))
	mux.HandleFunc("/healthz", HandlerReadiness)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", _PORT),
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %d\n", _ROOT, _PORT)
	log.Fatal(server.ListenAndServe())
}

func HandlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response body: %v", err)
	}
}
