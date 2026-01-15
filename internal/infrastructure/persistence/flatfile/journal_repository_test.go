package flatfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

func TestJournalRepositoryPersistsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryOne, err := journal.NewEntry(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueLove: "kindness",
	}, "note", "calm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(context.Background(), entryOne); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repoReloaded, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list, err := repoReloaded.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
	if list[0].Date.Format("2006-01-02") != "2024-02-01" {
		t.Fatalf("unexpected date: %s", list[0].Date.Format("2006-01-02"))
	}
}

func TestJournalRepositoryLatest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryOne, err := journal.NewEntry(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueLove: "kindness",
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entryTwo, err := journal.NewEntry(time.Date(2024, 2, 3, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueHappiness: "share",
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(context.Background(), entryOne); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(context.Background(), entryTwo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Date.Format("2006-01-02") != "2024-02-03" {
		t.Fatalf("unexpected latest date: %s", latest.Date.Format("2006-01-02"))
	}
}

func TestJournalRepositoryMultipleEntriesSameDate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryOne, err := journal.NewEntry(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueLove: "kindness",
	}, "first", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entryTwo, err := journal.NewEntry(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueHappiness: "share",
	}, "second", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(context.Background(), entryOne); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(context.Background(), entryTwo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repoReloaded, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := repoReloaded.Latest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Note != "second" {
		t.Fatalf("expected latest entry to be most recent, got %q", latest.Note)
	}

	list, err := repoReloaded.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	if list[0].Note != "first" || list[1].Note != "second" {
		t.Fatalf("expected entries in save order, got %q then %q", list[0].Note, list[1].Note)
	}
}

func TestJournalRepositoryLatestEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.Latest(context.Background()); err != journal.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDefaultJournalPathUsesXDGDataHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	path, err := DefaultJournalPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, "mt", "journal.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func TestDefaultJournalPathUsesHomeFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", dir)

	path, err := DefaultJournalPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, ".local", "share", "mt", "journal.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func TestNewJournalRepositoryRequiresPath(t *testing.T) {
	if _, err := NewJournalRepository(" "); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadIgnoresMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.List(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected missing file")
	}
}

func TestLoadIgnoresEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo, err := NewJournalRepository(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries, err := repo.List(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(entries))
	}
}

func TestNewJournalRepositoryFailsOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewJournalRepository(path); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewJournalRepositoryFailsOnInvalidDate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	data := []byte(`[
  {
    "date": "bad-date",
    "reflections": {
      "true-love": "note"
    }
  }
]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewJournalRepository(path); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewJournalRepositoryFailsOnInvalidEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	data := []byte(`[
  {
    "date": "2024-02-01",
    "reflections": {
      "unknown-precept": "note"
    }
  }
]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := NewJournalRepository(path); err == nil {
		t.Fatalf("expected error")
	}
}
