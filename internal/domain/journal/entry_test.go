package journal

import (
	"errors"
	"testing"
	"time"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name        string
		date        time.Time
		reflections map[Precept]string
		note        string
		mood        string
		foundation  Foundation
		timestamp   time.Time
		wantErr     error
		checkEntry  func(t *testing.T, entry *Entry)
	}{
		{
			name: "validates date",
			date: time.Time{},
			wantErr: ErrInvalidDate,
		},
		{
			name: "rejects unknown precept",
			date: time.Now(),
			reflections: map[Precept]string{
				Precept("unknown"): "reflection",
			},
			wantErr: ErrUnknownPrecept,
		},
		{
			name: "requires content",
			date: time.Now(),
			reflections: map[Precept]string{},
			note: "  ",
			mood: " ",
			wantErr: ErrEmptyEntry,
		},
		{
			name: "normalizes fields",
			date: time.Date(2024, 1, 2, 23, 30, 0, 0, time.FixedZone("local", -5*60*60)),
			reflections: map[Precept]string{
				ReverenceForLife: "  gratitude ",
				TrueLove:         " ",
			},
			note:       "  note ",
			mood:       " calm ",
			foundation: FoundationKaya,
			timestamp:  time.Date(2024, 1, 2, 12, 30, 0, 0, time.FixedZone("local", -5*60*60)),
			checkEntry: func(t *testing.T, entry *Entry) {
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
			},
		},
		{
			name: "defaults foundation",
			date: time.Now(),
			reflections: map[Precept]string{
				ReverenceForLife: "steady",
			},
			checkEntry: func(t *testing.T, entry *Entry) {
				if entry.Foundation != FoundationDhamma {
					t.Fatalf("expected default foundation, got %q", entry.Foundation)
				}
			},
		},
		{
			name: "rejects unknown foundation",
			date: time.Now(),
			reflections: map[Precept]string{
				ReverenceForLife: "steady",
			},
			foundation: Foundation("other"),
			wantErr: ErrUnknownFoundation,
		},
		{
			name: "defaults timestamp",
			date: time.Date(2024, 3, 10, 12, 0, 0, 0, time.FixedZone("local", -5*60*60)),
			reflections: map[Precept]string{
				ReverenceForLife: "steady",
			},
			foundation: FoundationDhamma,
			timestamp:  time.Time{},
			checkEntry: func(t *testing.T, entry *Entry) {
				if entry.Timestamp.Format(time.RFC3339) != "2024-03-10T00:00:00Z" {
					t.Fatalf("unexpected timestamp: %s", entry.Timestamp.Format(time.RFC3339))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := NewEntry(tt.date, tt.reflections, tt.note, tt.mood, tt.foundation, tt.timestamp)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkEntry != nil {
				tt.checkEntry(t, &entry)
			}
		})
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

func TestIsKnownFoundation(t *testing.T) {
	tests := []struct {
		foundation Foundation
		want       bool
	}{
		{FoundationKaya, true},
		{FoundationVedana, true},
		{FoundationCit, true},
		{FoundationDhamma, true},
		{Foundation("unknown"), false},
	}

	for _, tt := range tests {
		if got := IsKnownFoundation(tt.foundation); got != tt.want {
			t.Errorf("IsKnownFoundation(%q) = %v, want %v", tt.foundation, got, tt.want)
		}
	}
}

func TestParseFoundation(t *testing.T) {
	tests := []struct {
		input    string
		want     Foundation
		wantErr  bool
	}{
		{"", FoundationDhamma, false},
		{"d", FoundationDhamma, false},
		{"dhamma", FoundationDhamma, false},
		{"k", FoundationKaya, false},
		{"kaya", FoundationKaya, false},
		{"v", FoundationVedana, false},
		{"vedana", FoundationVedana, false},
		{"c", FoundationCit, false},
		{"cit", FoundationCit, false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		got, err := ParseFoundation(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseFoundation(%q) expected error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseFoundation(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFoundation(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestFoundationLabel(t *testing.T) {
	tests := []struct {
		foundation Foundation
		want       string
	}{
		{FoundationKaya, "Kaya"},
		{FoundationVedana, "Vedana"},
		{FoundationCit, "Cit"},
		{FoundationDhamma, "Dhamma"},
		{Foundation("unknown"), "unknown"},
	}

	for _, tt := range tests {
		if got := FoundationLabel(tt.foundation); got != tt.want {
			t.Errorf("FoundationLabel(%q) = %q, want %q", tt.foundation, got, tt.want)
		}
	}
}
