package memory

import (
	"context"
	"sort"
	"sync"

	"mindfulness/internal/domain/journal"
)

// JournalRepository is an in-memory implementation for journal entries.
type JournalRepository struct {
	mu      sync.RWMutex
	entries []journal.Entry
}

func NewJournalRepository() *JournalRepository {
	return &JournalRepository{
		entries: []journal.Entry{},
	}
}

func (r *JournalRepository) Save(_ context.Context, entry journal.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = append(r.entries, entry)
	return nil
}

func (r *JournalRepository) Latest(_ context.Context) (*journal.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.entries) == 0 {
		return nil, journal.ErrNotFound
	}

	latestIndex := 0
	for i := 1; i < len(r.entries); i++ {
		current := r.entries[i].Date
		latest := r.entries[latestIndex].Date
		if current.After(latest) || current.Equal(latest) {
			latestIndex = i
		}
	}
	latest := r.entries[latestIndex]
	copy := latest
	return &copy, nil
}

func (r *JournalRepository) List(_ context.Context) ([]journal.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.entries) == 0 {
		return nil, nil
	}

	entries := append([]journal.Entry{}, r.entries...)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})
	return entries, nil
}
