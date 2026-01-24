package journal

import (
	"errors"
	"testing"
	"time"
)

func TestNewEntryValidatesDate(t *testing.T) {
	_, err := NewEntry(time.Time{}, nil, "note", "calm", FoundationDhamma)
	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

func TestNewEntryRejectsUnknownPrecept(t *testing.T) {
	_, err := NewEntry(time.Now(), map[Precept]string{
		Precept("unknown"): "reflection",
	}, "", "", FoundationDhamma)
	if !errors.Is(err, ErrUnknownPrecept) {
		t.Fatalf("expected ErrUnknownPrecept, got %v", err)
	}
}

func TestNewEntryRequiresContent(t *testing.T) {
	_, err := NewEntry(time.Now(), map[Precept]string{}, "  ", " ", FoundationDhamma)
	if !errors.Is(err, ErrEmptyEntry) {
		t.Fatalf("expected ErrEmptyEntry, got %v", err)
	}
}

func TestNewEntryNormalizesFields(t *testing.T) {
	date := time.Date(2024, 1, 2, 23, 30, 0, 0, time.FixedZone("local", -5*60*60))
	entry, err := NewEntry(date, map[Precept]string{
		ReverenceForLife: "  gratitude ",
		TrueLove:         " ",
	}, "  note ", " calm ", FoundationKaya)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Date.Format("2006-01-02") != "2024-01-03" {
		t.Fatalf("expected normalized date, got %s", entry.Date.Format("2006-01-02"))
	}
	if entry.Note != "note" {
		t.Fatalf("expected trimmed note, got %q", entry.Note)
	}
	if entry.Mood != "calm" {
		t.Fatalf("expected trimmed mood, got %q", entry.Mood)
	}
	if len(entry.Reflections) != 1 {
		t.Fatalf("expected 1 reflection, got %d", len(entry.Reflections))
	}
	if entry.Reflections[ReverenceForLife] != "gratitude" {
		t.Fatalf("expected trimmed reflection")
	}
	if entry.Foundation != FoundationKaya {
		t.Fatalf("expected foundation to be set")
	}
}

func TestNewEntryDefaultsFoundation(t *testing.T) {
	entry, err := NewEntry(time.Now(), map[Precept]string{
		ReverenceForLife: "steady",
	}, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Foundation != FoundationDhamma {
		t.Fatalf("expected default foundation, got %q", entry.Foundation)
	}
}

func TestNewEntryRejectsUnknownFoundation(t *testing.T) {
	_, err := NewEntry(time.Now(), map[Precept]string{
		ReverenceForLife: "steady",
	}, "", "", Foundation("other"))
	if !errors.Is(err, ErrUnknownFoundation) {
		t.Fatalf("expected ErrUnknownFoundation, got %v", err)
	}
}

func TestSortedPrecepts(t *testing.T) {
	entry := Entry{
		Reflections: map[Precept]string{
			TrueLove:         "a",
			TrueHappiness:    "b",
			ReverenceForLife: "c",
		},
	}
	sorted := entry.SortedPrecepts()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 precepts, got %d", len(sorted))
	}
	if sorted[0] != ReverenceForLife || sorted[1] != TrueHappiness || sorted[2] != TrueLove {
		t.Fatalf("unexpected order: %v", sorted)
	}
}
