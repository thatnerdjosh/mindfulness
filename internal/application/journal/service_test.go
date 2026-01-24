package journal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

type fakeRepo struct {
	saved   journal.Entry
	entries []journal.Entry
	err     error
}

func (f *fakeRepo) Save(_ context.Context, entry journal.Entry) error {
	if f.err != nil {
		return f.err
	}
	f.saved = entry
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeRepo) Latest(_ context.Context) (*journal.Entry, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.entries) == 0 {
		return nil, journal.ErrNotFound
	}
	latest := f.entries[len(f.entries)-1]
	return &latest, nil
}

func (f *fakeRepo) List(_ context.Context) ([]journal.Entry, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]journal.Entry{}, f.entries...), nil
}

func TestRecordEntrySavesNormalizedEntry(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	date := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	entry, err := svc.RecordEntry(context.Background(), date, map[journal.Precept]string{
		journal.ReverenceForLife: " clarity ",
	}, " note ", " calm ", journal.FoundationDhamma)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Note != "note" {
		t.Fatalf("expected trimmed note, got %q", entry.Note)
	}
	if repo.saved.Date.IsZero() {
		t.Fatalf("expected entry saved")
	}
	if repo.saved.Reflections[journal.ReverenceForLife] != "clarity" {
		t.Fatalf("expected trimmed reflection")
	}
}

func TestRecordEntryPropagatesErrors(t *testing.T) {
	expected := errors.New("save failed")
	repo := &fakeRepo{err: expected}
	svc := NewService(repo)

	_, err := svc.RecordEntry(context.Background(), time.Now(), map[journal.Precept]string{
		journal.ReverenceForLife: "note",
	}, "", "", journal.FoundationDhamma)
	if !errors.Is(err, expected) {
		t.Fatalf("expected error to propagate, got %v", err)
	}
}

func TestRecordEntryValidatesEntry(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.RecordEntry(context.Background(), time.Now(), map[journal.Precept]string{}, " ", " ", journal.FoundationDhamma)
	if !errors.Is(err, journal.ErrEmptyEntry) {
		t.Fatalf("expected ErrEmptyEntry, got %v", err)
	}
}

func TestLatestAndListDelegateToRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	entry, err := journal.NewEntry(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueHappiness: "share",
	}, "", "", journal.FoundationDhamma)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.entries = append(repo.entries, entry)
	latest, err := svc.LatestEntry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Date.Format("2006-01-02") != "2024-01-02" {
		t.Fatalf("unexpected latest date: %s", latest.Date.Format("2006-01-02"))
	}

	list, err := svc.ListEntries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
}
