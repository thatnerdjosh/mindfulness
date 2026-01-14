package journal

import "time"

// Adherence tracks whether each precept is currently being kept.
type Adherence map[Precept]bool

// DefaultAdherence returns adherence with every known precept set to true.
func DefaultAdherence() Adherence {
	adherence := make(Adherence, len(AllPrecepts()))
	for _, info := range AllPrecepts() {
		adherence[info.ID] = true
	}
	return adherence
}

// AdherenceLogEntry captures a change in adherence for a precept.
type AdherenceLogEntry struct {
	At      time.Time
	Precept Precept
	From    bool
	To      bool
	Note    string
}
