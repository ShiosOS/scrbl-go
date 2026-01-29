package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/juliuswalton/scrbl-server/store"
)

// Server is the HTTP API server.
type Server struct {
	store  *store.Store
	apiKey string
	mux    *http.ServeMux
}

// New creates a new API server.
func New(s *store.Store, apiKey string) *Server {
	srv := &Server{
		store:  s,
		apiKey: apiKey,
		mux:    http.NewServeMux(),
	}
	srv.routes()
	return srv
}

// Handler returns the http.Handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/notes", s.auth(s.handleNotesList))
	s.mux.HandleFunc("/api/notes/", s.auth(s.handleNotesItem))
	s.mux.HandleFunc("/api/search", s.auth(s.handleSearch))
	s.mux.HandleFunc("/health", s.handleHealth)
}

// --- Middleware ---

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" {
			next(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token != s.apiKey {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// --- Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GET /api/notes — list all dates
func (s *Server) handleNotesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for ?since= param for incremental sync
	since := r.URL.Query().Get("since")
	if since != "" {
		notes, err := s.store.GetUpdatedSince(since)
		if err != nil {
			log.Printf("ERROR list updated since: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, notes)
		return
	}

	dates, err := s.store.ListDates()
	if err != nil {
		log.Printf("ERROR list dates: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, dates)
}

// GET/PUT /api/notes/:date
func (s *Server) handleNotesItem(w http.ResponseWriter, r *http.Request) {
	// Extract date from path: /api/notes/2025-01-29
	date := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if date == "" || len(date) != 10 {
		http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getNoteByDate(w, date)
	case http.MethodPut:
		s.putNoteByDate(w, r, date)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getNoteByDate(w http.ResponseWriter, date string) {
	note, err := s.store.Get(date)
	if err != nil {
		log.Printf("ERROR get note %s: %v", date, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if note == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{
		"date":    note.Date,
		"content": note.Content,
	})
}

type putNoteRequest struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

func (s *Server) putNoteByDate(w http.ResponseWriter, r *http.Request, date string) {
	var req putNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Use the URL date as the source of truth
	note, err := s.store.Upsert(date, req.Content)
	if err != nil {
		log.Printf("ERROR upsert note %s: %v", date, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, note)
}

// GET /api/search?q=query
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	results, err := s.store.Search(q)
	if err != nil {
		log.Printf("ERROR search: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, results)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ERROR encoding json: %v", err)
	}
}
