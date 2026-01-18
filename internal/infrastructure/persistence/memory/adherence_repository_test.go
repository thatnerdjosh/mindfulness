package memory

import (
	"context"
	"testing"

	"github.com/thatnerdjosh/mindfulness/internal/domain/adherence"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

func TestAdherenceRepositoryDefaults(t *testing.T) {
	repo := NewAdherenceRepository()
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

func TestAdherenceRepositorySave(t *testing.T) {
	repo := NewAdherenceRepository()
	state := adherence.DefaultAdherence()
	state[journal.TrueHappiness] = false
	if err := repo.Save(context.Background(), state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	loaded, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded[journal.TrueHappiness] {
		t.Fatalf("expected TrueHappiness false")
	}
}
