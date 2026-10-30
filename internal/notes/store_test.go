package notes

import (
	"context"
	"strings"
	"testing"
)

func TestCleanText(t *testing.T) {
	for _, text := range []string{"", " \t\n", "hello\x00world", strings.Repeat("я", 2001), string([]byte{0xff})} {
		if _, err := CleanText(text); err != ErrInvalidText {
			t.Fatalf("expected invalid text, got %v", err)
		}
	}
	if text, err := CleanText("  hello  "); text != "hello" || err != nil {
		t.Fatalf("%q %v", text, err)
	}
	if _, err := CleanText(strings.Repeat("я", 2000)); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidConfigurationDoesNotLeakCredentials(t *testing.T) {
	_, err := Open(context.Background(), "postgres://user:private-password@127.0.0.1:invalid/database")
	if err == nil || strings.Contains(err.Error(), "private-password") {
		t.Fatalf("unexpected configuration error: %v", err)
	}
}
