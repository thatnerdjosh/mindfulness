package adherence

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/adherence"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

type fakeAdherenceRepo struct {
	adherence adherence.Adherence
	log       []adherence.AdherenceLogEntry
	err       error
}

func (f *fakeAdherenceRepo) Get(_ context.Context) (adherence.Adherence, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.adherence, nil
}

func (f *fakeAdherenceRepo) Save(_ context.Context, a adherence.Adherence) error {
	if f.err != nil {
		return f.err
	}
	f.adherence = a
	return nil
}

func (f *fakeAdherenceRepo) AppendLog(_ context.Context, entry adherence.AdherenceLogEntry) error {
	if f.err != nil {
		return f.err
	}
	f.log = append(f.log, entry)
	return nil
}

func TestServiceCurrent(t *testing.T) {
	repo := &fakeAdherenceRepo{adherence: adherence.DefaultAdherence()}
	svc := NewService(repo)

	current, err := svc.Current(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(current) != len(journal.AllPrecepts()) {
		t.Fatalf("expected %d precepts, got %d", len(journal.AllPrecepts()), len(current))
	}
}

func TestServiceSet(t *testing.T) {
	tests := []struct {
		name      string
		current   adherence.Adherence
		next      adherence.Adherence
		notes     map[journal.Precept]string
		repoErr   error
		wantErr   string
		check     func(t *testing.T, repo *fakeAdherenceRepo, now time.Time)
	}{
		{
			name:    "set with changes",
			current: adherence.Adherence{journal.TrueLove: true, journal.TrueHappiness: true},
			next:    adherence.Adherence{journal.TrueLove: false},
			notes:   map[journal.Precept]string{journal.TrueLove: " slipped "},
			check: func(t *testing.T, repo *fakeAdherenceRepo, now time.Time) {
				if repo.adherence[journal.TrueLove] {
					t.Fatalf("expected TrueLove false")
				}
				if !repo.adherence[journal.TrueHappiness] {
					t.Fatalf("expected TrueHappiness unchanged true")
				}
				if len(repo.log) != 1 {
					t.Fatalf("expected 1 log entry, got %d", len(repo.log))
				}
				entry := repo.log[0]
				if entry.Precept != journal.TrueLove || entry.From != true || entry.To != false || entry.Note != "slipped" {
					t.Fatalf("unexpected log entry: %+v", entry)
				}
				if entry.At != now {
					t.Fatalf("expected log at %v, got %v", now, entry.At)
				}
			},
		},
		{
			name:    "unknown precept",
			current: adherence.Adherence{journal.TrueLove: true},
			next:    adherence.Adherence{journal.Precept("unknown"): true},
			wantErr: "unknown precept",
		},
		{
			name:    "save error",
			current: adherence.Adherence{journal.TrueLove: true},
			next:    adherence.Adherence{journal.TrueLove: false},
			repoErr: errors.New("save failed"),
			wantErr: "save failed",
		},
		{
			name:    "no changes",
			current: adherence.Adherence{journal.TrueLove: true},
			next:    adherence.Adherence{journal.TrueLove: true},
			check: func(t *testing.T, repo *fakeAdherenceRepo, now time.Time) {
				if len(repo.log) != 0 {
					t.Fatalf("expected no log entries, got %d", len(repo.log))
				}
			},
		},
		{
			name:    "with nil notes",
			current: adherence.Adherence{journal.TrueLove: true},
			next:    adherence.Adherence{journal.TrueLove: false},
			notes:   nil,
			check: func(t *testing.T, repo *fakeAdherenceRepo, now time.Time) {
				if len(repo.log) != 1 {
					t.Fatalf("expected 1 log entry, got %d", len(repo.log))
				}
				entry := repo.log[0]
				if entry.Note != "" {
					t.Fatalf("expected empty note, got %q", entry.Note)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAdherenceRepo{adherence: tt.current, err: tt.repoErr}
			now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			svc := &Service{repo: repo, now: func() time.Time { return now }}

			err := svc.Set(context.Background(), tt.next, tt.notes)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, repo, now)
			}
		})
	}
}