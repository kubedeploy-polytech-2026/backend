package api

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

type versionResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

func NewHandler(store NotesStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})
	mux.HandleFunc("GET /readyz", ready(store))
	mux.HandleFunc("GET /api/notes", listNotes(store))
	mux.HandleFunc("POST /api/notes", createNote(store))
	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, versionResponse{
			Service: "kubedeploy-backend",
			Version: "0.2.0",
		})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
