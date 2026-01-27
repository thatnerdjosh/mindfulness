package flatfile

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/adherence"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
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

	state := adherence.DefaultAdherence()
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

	entry := adherence.AdherenceLogEntry{
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

func TestNewAdherenceRepository(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		logPath string
		setup   func(t *testing.T, dir string) string
		wantErr bool
	}{
		{
			name:    "requires path",
			path:    "",
			logPath: "log",
			wantErr: true,
		},
		{
			name:    "requires log path",
			path:    "path",
			logPath: " ",
			wantErr: true,
		},
		{
			name: "fails on invalid JSON",
			setup: func(t *testing.T, dir string) string {
				path := filepath.Join(dir, "adherence.json")
				if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "fails on unknown precept",
			setup: func(t *testing.T, dir string) string {
				path := filepath.Join(dir, "adherence.json")
				data := []byte(`{
  "unknown-precept": true
}`)
				if err := os.WriteFile(path, data, 0o600); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := tt.path
			logPath := tt.logPath
			if tt.setup != nil {
				path = tt.setup(t, dir)
				logPath = filepath.Join(dir, "adherence.log.jsonl")
			}
			_, err := NewAdherenceRepository(path, logPath)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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
