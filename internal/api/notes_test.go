package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kubedeploy/backend/internal/notes"
)

type fakeNotes struct {
	err     error
	created string
	items   []notes.Note
	calls   int
}

func (s *fakeNotes) Ready(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return s.err
}
func (s *fakeNotes) List(ctx context.Context) ([]notes.Note, error) {
	return s.items, s.Ready(ctx)
}
func (s *fakeNotes) Create(ctx context.Context, text string) (notes.Note, error) {
	s.calls++
	s.created = text
	return notes.Note{ID: 1, Text: text, CreatedAt: time.Unix(1, 0).UTC()}, s.Ready(ctx)
}

func TestCreateNoteValidation(t *testing.T) {
	for _, tc := range []struct {
		name, body, contentType string
		status                  int
	}{
		{"missing media type", `{"text":"hello"}`, "", 415},
		{"invalid media type", `{"text":"hello"}`, "text/plain", 415},
		{"malformed", `{`, "application/json", 400},
		{"empty body", ``, "application/json", 400},
		{"unknown field", `{"text":"hello","extra":1}`, "application/json", 400},
		{"array", `[]`, "application/json", 400},
		{"null", `null`, "application/json", 400},
		{"second object", `{"text":"hello"}{}`, "application/json", 400},
		{"trailing garbage", `{"text":"hello"}bad`, "application/json", 400},
		{"empty text", `{"text":" \n "}`, "application/json", 400},
		{"long text", `{"text":"` + strings.Repeat("я", 2001) + `"}`, "application/json", 400},
		{"large body", `{"text":"` + strings.Repeat("x", 20*1024) + `"}`, "application/json", 413},
		{"valid", `{"text":" hello "}`, "application/json; charset=utf-8", 201},
		{"unicode boundary", `{"text":"` + strings.Repeat("я", 2000) + `"}`, "application/json", 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeNotes{}
			request := httptest.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", tc.contentType)
			response := httptest.NewRecorder()
			NewHandler(store).ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status != 201 && store.calls != 0 {
				t.Fatal("invalid input reached storage")
			}
			if tc.status == 201 {
				var note notes.Note
				if err := json.Unmarshal(response.Body.Bytes(), &note); err != nil || note.ID != 1 || note.Text != store.created || note.CreatedAt.IsZero() {
					t.Fatalf("invalid created note: %s", response.Body.String())
				}
				if tc.name == "valid" && note.Text != "hello" {
					t.Fatal("text was not trimmed")
				}
			}
		})
	}
}

func TestDatabaseReadinessAndFailure(t *testing.T) {
	for _, store := range []NotesStore{nil, &fakeNotes{err: errors.New("private-database-password")}, &fakeNotes{}} {
		for _, path := range []string{"/health", "/readyz", "/api/notes"} {
			response := httptest.NewRecorder()
			NewHandler(store).ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			want := http.StatusOK
			if path != "/health" && (store == nil || store.(*fakeNotes).err != nil) {
				want = http.StatusServiceUnavailable
			}
			if response.Code != want || strings.Contains(response.Body.String(), "private-database-password") {
				t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
			}
		}
	}
}

func TestCreateUnavailableAndCanceled(t *testing.T) {
	for _, store := range []NotesStore{nil, &fakeNotes{err: errors.New("private-database-password")}, &fakeNotes{}} {
		request := httptest.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(`{"text":"hello"}`))
		request.Header.Set("Content-Type", "application/json")
		ctx, cancel := context.WithCancel(request.Context())
		cancel()
		response := httptest.NewRecorder()
		NewHandler(store).ServeHTTP(response, request.WithContext(ctx))
		if response.Code != 503 || strings.Contains(response.Body.String(), "private-database-password") {
			t.Fatalf("%d %s", response.Code, response.Body.String())
		}
	}
}

func TestEmptyNotesListAndUnsupportedMethod(t *testing.T) {
	handler := NewHandler(&fakeNotes{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/notes", nil))
	if response.Code != 200 || response.Body.String() != "[]\n" {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/notes", nil))
	if response.Code != 405 {
		t.Fatalf("unexpected status %d", response.Code)
	}
}
