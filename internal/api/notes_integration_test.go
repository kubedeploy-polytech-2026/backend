package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"kubedeploy/backend/internal/notes"
)

func TestPostgresNotesPersistence(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cleanup, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal("invalid integration database configuration")
	}
	defer cleanup.Close()
	store, err := notes.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	defer func() {
		store.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cleanupCancel()
		for _, id := range ids {
			if _, err := cleanup.Exec(cleanupCtx, "DELETE FROM app.notes WHERE id = $1", id); err != nil {
				t.Error("could not remove integration test note")
			}
		}
	}()
	create := func(text string) notes.Note {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"text": text})
		request := httptest.NewRequest(http.MethodPost, "/api/notes", strings.NewReader(string(body)))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		NewHandler(store).ServeHTTP(response, request)
		if response.Code != 201 {
			t.Fatalf("create status %d: %s", response.Code, response.Body.String())
		}
		var note notes.Note
		if err := json.Unmarshal(response.Body.Bytes(), &note); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, note.ID)
		if note.ID <= 0 || note.Text != text || note.CreatedAt.IsZero() {
			t.Fatal("invalid persisted note")
		}
		return note
	}
	first := create("integration: Robert'); DROP TABLE app.notes; -- привет")
	second := create("integration: survives pool restart")
	store.Close()
	response := httptest.NewRecorder()
	NewHandler(store).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != 503 {
		t.Fatalf("closed database readiness: %d", response.Code)
	}
	reopened, err := notes.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	store = reopened
	response = httptest.NewRecorder()
	NewHandler(store).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/notes", nil))
	if response.Code != 200 {
		t.Fatalf("list status %d", response.Code)
	}
	var result []notes.Note
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	found := make(map[int64]notes.Note)
	for i, note := range result {
		if i > 0 && note.ID >= result[i-1].ID {
			t.Fatal("notes are not ordered newest first")
		}
		found[note.ID] = note
	}
	for _, want := range []notes.Note{first, second} {
		got := found[want.ID]
		if got.Text != want.Text || !got.CreatedAt.Equal(want.CreatedAt) {
			t.Fatal("note changed after reopening PostgreSQL pool")
		}
	}
	response = httptest.NewRecorder()
	NewHandler(store).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != 200 {
		t.Fatalf("recovered database readiness: %d", response.Code)
	}
}
