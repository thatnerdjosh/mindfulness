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

func TestRecordEntry(t *testing.T) {
	saveFailedErr := errors.New("save failed")
	tests := []struct {
		name        string
		date        time.Time
		reflections map[journal.Precept]string
		note        string
		mood        string
		foundation  journal.Foundation
		repoErr     error
		wantErr     error
		check       func(t *testing.T, entry *journal.Entry, repo *fakeRepo)
	}{
		{
			name: "saves normalized entry",
			date: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			reflections: map[journal.Precept]string{
				journal.ReverenceForLife: " clarity ",
			},
			note:       " note ",
			mood:       " calm ",
			foundation: journal.FoundationDhamma,
			check: func(t *testing.T, entry *journal.Entry, repo *fakeRepo) {
				if entry.Note != "note" {
					t.Fatalf("expected trimmed note, got %q", entry.Note)
				}
				if repo.saved.Date.IsZero() {
					t.Fatalf("expected entry saved")
				}
				if repo.saved.Reflections[journal.ReverenceForLife] != "clarity" {
					t.Fatalf("expected trimmed reflection")
				}
			},
		},
		{
			name: "propagates errors",
			date: time.Now(),
			reflections: map[journal.Precept]string{
				journal.ReverenceForLife: "note",
			},
			foundation: journal.FoundationDhamma,
			repoErr:   saveFailedErr,
			wantErr:   saveFailedErr,
		},
		{
			name: "validates entry",
			date: time.Now(),
			reflections: map[journal.Precept]string{},
			note:        " ",
			mood:        " ",
			foundation:  journal.FoundationDhamma,
			wantErr:     journal.ErrEmptyEntry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{err: tt.repoErr}
			svc := NewService(repo)

			entry, err := svc.RecordEntry(context.Background(), tt.date, tt.reflections, tt.note, tt.mood, tt.foundation)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, &entry, repo)
			}
		})
	}
}

func TestLatestAndListDelegateToRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	entry, err := journal.NewEntry(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), map[journal.Precept]string{
		journal.TrueHappiness: "share",
	}, "", "", journal.FoundationDhamma, time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC))
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
