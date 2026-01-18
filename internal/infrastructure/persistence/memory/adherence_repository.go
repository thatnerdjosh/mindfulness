package memory

import (
	"context"
	"sync"

	"github.com/thatnerdjosh/mindfulness/internal/domain/adherence"
)

// AdherenceRepository is an in-memory implementation for adherence state and logs.
type AdherenceRepository struct {
	mu        sync.RWMutex
	adherence adherence.Adherence
	logs      []adherence.AdherenceLogEntry
}

func NewAdherenceRepository() *AdherenceRepository {
	return &AdherenceRepository{
		adherence: adherence.DefaultAdherence(),
		logs:      []adherence.AdherenceLogEntry{},
	}
}

func (r *AdherenceRepository) Get(_ context.Context) (adherence.Adherence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(adherence.Adherence, len(r.adherence))
	for precept, value := range r.adherence {
		copy[precept] = value
	}
	return copy, nil
}

func (r *AdherenceRepository) Save(_ context.Context, state adherence.Adherence) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := make(adherence.Adherence, len(state))
	for precept, value := range state {
		copy[precept] = value
	}
	r.adherence = copy
	return nil
}

func (r *AdherenceRepository) AppendLog(_ context.Context, entry adherence.AdherenceLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = append(r.logs, entry)
	return nil
}
