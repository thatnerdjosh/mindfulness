package flatfile

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mindfulness/internal/domain/journal"
)

func TestAdherenceRepositoryDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adherence.json")
	logPath := filepath.Join(dir, "adherence.log.jsonl")

	repo, err := NewAdherenceRepository(path, logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, info := range journal.AllPrecepts() {
		if value, ok := state[info.ID]; !ok || !value {
			t.Fatalf("expected default true for %s", info.ID)
		}
	}
}

func TestAdherenceRepositoryPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adherence.json")
	logPath := filepath.Join(dir, "adherence.log.jsonl")

	repo, err := NewAdherenceRepository(path, logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state := journal.DefaultAdherence()
	state[journal.TrueLove] = false
	if err := repo.Save(context.Background(), state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repoReloaded, err := NewAdherenceRepository(path, logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := repoReloaded.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded[journal.TrueLove] {
		t.Fatalf("expected TrueLove false after reload")
	}
}

func TestAdherenceRepositoryAppendLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adherence.json")
	logPath := filepath.Join(dir, "adherence.log.jsonl")

	repo, err := NewAdherenceRepository(path, logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := journal.AdherenceLogEntry{
		At:      time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC),
		Precept: journal.TrueLove,
		From:    true,
		To:      false,
		Note:    "slipped",
	}
	if err := repo.AppendLog(context.Background(), entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var record adherenceLogRecord
	if err := json.Unmarshal(bytesTrimSpace(data), &record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.Precept != "true-love" || record.From != true || record.To != false {
		t.Fatalf("unexpected log record: %+v", record)
	}
}

func TestNewAdherenceRepositoryRequiresPath(t *testing.T) {
	if _, err := NewAdherenceRepository("", "log"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := NewAdherenceRepository("path", " "); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewAdherenceRepositoryFailsOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adherence.json")
	logPath := filepath.Join(dir, "adherence.log.jsonl")

	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewAdherenceRepository(path, logPath); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewAdherenceRepositoryFailsOnUnknownPrecept(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adherence.json")
	logPath := filepath.Join(dir, "adherence.log.jsonl")

	data := []byte(`{
  "unknown-precept": true
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewAdherenceRepository(path, logPath); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDefaultAdherencePathUsesXDGDataHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	path, err := DefaultAdherencePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, "mt", "adherence.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func TestDefaultAdherencePathUsesHomeFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", dir)

	path, err := DefaultAdherencePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, ".local", "share", "mt", "adherence.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func bytesTrimSpace(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r' || data[len(data)-1] == ' ') {
		data = data[:len(data)-1]
	}
	return data
}
