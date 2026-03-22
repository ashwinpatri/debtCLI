package git

import (
	"context"
	"testing"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

func TestParsePorcelain_Basic(t *testing.T) {
	// Minimal porcelain output for a single-line file.
	input := "abcdef1234567890abcdef1234567890abcdef12 1 1 1\n" +
		"author Jane Doe\n" +
		"author-mail <jane@example.com>\n" +
		"author-time 1700000000\n" +
		"author-tz +0000\n" +
		"committer Jane Doe\n" +
		"committer-mail <jane@example.com>\n" +
		"committer-time 1700000000\n" +
		"committer-tz +0000\n" +
		"summary initial\n" +
		"filename main.go\n" +
		"\tpackage main\n"

	result, err := parsePorcelain([]byte(input))
	if err != nil {
		t.Fatalf("parsePorcelain: %v", err)
	}
	info, ok := result[1]
	if !ok {
		t.Fatal("expected entry for line 1")
	}
	if info.Author != "Jane Doe" {
		t.Errorf("author: got %q", info.Author)
	}
	if info.AuthorEmail != "jane@example.com" {
		t.Errorf("author-mail: got %q", info.AuthorEmail)
	}
	if info.Timestamp != 1700000000 {
		t.Errorf("timestamp: got %d", info.Timestamp)
	}
}

func TestCountLines(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"abc123 first commit\n", 1},
		{"abc123 first\ndef456 second\n", 2},
		{"abc123 first\n\ndef456 second\n", 2}, // blank lines not counted
	}
	for _, tc := range cases {
		got := countLines([]byte(tc.input))
		if got != tc.want {
			t.Errorf("countLines(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestCache_BlameReadWrite(t *testing.T) {
	c := newCache()
	data := map[int]BlameInfo{1: {Author: "ashwin"}}
	c.setBlame("file.go", data, nil)

	entry, ok := c.getBlame("file.go")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if entry.result[1].Author != "ashwin" {
		t.Errorf("cached author: got %q", entry.result[1].Author)
	}
}

func TestCache_ChurnReadWrite(t *testing.T) {
	c := newCache()
	c.setChurn("file.go", 7, nil)

	entry, ok := c.getChurn("file.go")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if entry.result != 7 {
		t.Errorf("cached churn: got %d", entry.result)
	}
}

func TestCache_Miss(t *testing.T) {
	c := newCache()
	_, ok := c.getBlame("missing.go")
	if ok {
		t.Error("expected cache miss")
	}
}

// TestMockClient_BlameIntegration verifies the mock integrates with models correctly.
func TestMockClient_BlameIntegration(t *testing.T) {
	mc := NewMockClient()
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	mc.BlameData["main.go"] = map[int]BlameInfo{
		1: {Author: "ashwin", AuthorEmail: "a@example.com", Timestamp: ts},
	}

	result, err := mc.Blame(context.Background(), "/repo", "main.go")
	if err != nil {
		t.Fatalf("Blame: %v", err)
	}

	info := result[1]
	item := models.DebtItem{
		Author:      info.Author,
		AuthorEmail: info.AuthorEmail,
		Date:        time.Unix(info.Timestamp, 0),
	}
	if item.Author != "ashwin" {
		t.Errorf("author: got %q", item.Author)
	}
}
