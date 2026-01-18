package adherence

import (
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

// Adherence tracks whether each precept is currently being kept.
type Adherence map[journal.Precept]bool

// DefaultAdherence returns adherence with every known precept set to true.
func DefaultAdherence() Adherence {
	adherence := make(Adherence, len(journal.AllPrecepts()))
	for _, info := range journal.AllPrecepts() {
		adherence[info.ID] = true
	}
	return adherence
}

// AdherenceLogEntry captures a change in adherence for a precept.
type AdherenceLogEntry struct {
	At      time.Time
	Precept journal.Precept
	From    bool
	To      bool
	Note    string
}
