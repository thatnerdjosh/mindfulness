package memory

import (
	"context"
	"testing"
	"time"

	"mindfulness/internal/domain/journal"
)

func TestJournalRepositoryLatestAndList(t *testing.T) {
	repo := NewJournalRepository()

	entryOne, err := journal.NewEntry(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueLove: "kindness",
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entryTwo, err := journal.NewEntry(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.ReverenceForLife: "care",
	}, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(context.Background(), entryTwo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(context.Background(), entryOne); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Date.Format("2006-01-02") != "2024-01-03" {
		t.Fatalf("expected latest date 2024-01-03, got %s", latest.Date.Format("2006-01-02"))
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	if list[0].Date.Format("2006-01-02") != "2024-01-01" {
		t.Fatalf("expected first entry 2024-01-01, got %s", list[0].Date.Format("2006-01-02"))
	}
	if list[1].Date.Format("2006-01-02") != "2024-01-03" {
		t.Fatalf("expected second entry 2024-01-03, got %s", list[1].Date.Format("2006-01-02"))
	}
}

func TestJournalRepositoryMultipleEntriesSameDate(t *testing.T) {
	repo := NewJournalRepository()

	entryOne, err := journal.NewEntry(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueLove: "kindness",
	}, "first", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entryTwo, err := journal.NewEntry(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
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

	latest, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Note != "second" {
		t.Fatalf("expected latest entry to be most recent, got %q", latest.Note)
	}

	list, err := repo.List(context.Background())
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

func TestJournalRepositoryEmpty(t *testing.T) {
	repo := NewJournalRepository()
	latest, err := repo.Latest(context.Background())
	if latest != nil {
		t.Fatalf("expected nil latest")
	}
	if err != journal.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list")
	}
}
