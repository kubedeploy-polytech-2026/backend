package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"kubedeploy/backend/internal/notes"
)

type NotesStore interface {
	Ready(context.Context) error
	List(context.Context) ([]notes.Note, error)
	Create(context.Context, string) (notes.Note, error)
}

func apiError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func ready(store NotesStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if store == nil || store.Ready(ctx) != nil {
			apiError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	}
}

func listNotes(store NotesStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			apiError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		result, err := store.List(ctx)
		if err != nil {
			apiError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		if result == nil {
			result = []notes.Note{}
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func createNote(store NotesStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			apiError(w, http.StatusUnsupportedMediaType, "expected application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var input struct {
			Text string `json:"text"`
		}
		err = decoder.Decode(&input)
		if err == nil {
			var extra any
			if end := decoder.Decode(&extra); end != io.EOF {
				if end == nil {
					err = errors.New("multiple JSON values")
				} else {
					err = end
				}
			}
		}
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				apiError(w, http.StatusRequestEntityTooLarge, "request body too large")
			} else {
				apiError(w, http.StatusBadRequest, "invalid JSON")
			}
			return
		}
		text, err := notes.CleanText(input.Text)
		if err != nil {
			apiError(w, http.StatusBadRequest, notes.ErrInvalidText.Error())
			return
		}
		if store == nil {
			apiError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		note, err := store.Create(ctx, text)
		if err != nil {
			apiError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		writeJSON(w, http.StatusCreated, note)
	}
}
