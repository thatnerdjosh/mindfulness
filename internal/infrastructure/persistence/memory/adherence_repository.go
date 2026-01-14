package memory

import (
	"context"
	"sync"

	"mindfulness/internal/domain/journal"
)

// AdherenceRepository is an in-memory implementation for adherence state and logs.
type AdherenceRepository struct {
	mu        sync.RWMutex
	adherence journal.Adherence
	logs      []journal.AdherenceLogEntry
}

func NewAdherenceRepository() *AdherenceRepository {
	return &AdherenceRepository{
		adherence: journal.DefaultAdherence(),
		logs:      []journal.AdherenceLogEntry{},
	}
}

func (r *AdherenceRepository) Get(_ context.Context) (journal.Adherence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(journal.Adherence, len(r.adherence))
	for precept, value := range r.adherence {
		copy[precept] = value
	}
	return copy, nil
}

func (r *AdherenceRepository) Save(_ context.Context, adherence journal.Adherence) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := make(journal.Adherence, len(adherence))
	for precept, value := range adherence {
		copy[precept] = value
	}
	r.adherence = copy
	return nil
}

func (r *AdherenceRepository) AppendLog(_ context.Context, entry journal.AdherenceLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = append(r.logs, entry)
	return nil
}
