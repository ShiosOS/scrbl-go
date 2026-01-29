package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/juliuswalton/scrbl-server/api"
	"github.com/juliuswalton/scrbl-server/store"
)

func main() {
	port := flag.String("port", envOr("PORT", "8080"), "server port")
	dbPath := flag.String("db", envOr("DB_PATH", "./scrbl.db"), "SQLite database path")
	apiKey := flag.String("api-key", envOr("API_KEY", ""), "API key for authentication (empty = no auth)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Open database
	s, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer s.Close()

	// Create API server
	srv := api.New(s, *apiKey)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("scrbl-server listening on %s", addr)
	if *apiKey != "" {
		log.Printf("API key authentication enabled")
	} else {
		log.Printf("WARNING: no API key set, server is unauthenticated")
	}

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
