package adherence

import (
	"testing"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

func TestDefaultAdherence(t *testing.T) {
	adherence := DefaultAdherence()
	if len(adherence) != len(journal.AllPrecepts()) {
		t.Fatalf("expected %d precepts, got %d", len(journal.AllPrecepts()), len(adherence))
	}
	for _, info := range journal.AllPrecepts() {
		if value, ok := adherence[info.ID]; !ok || !value {
			t.Fatalf("expected default true for %s", info.ID)
		}
	}
}