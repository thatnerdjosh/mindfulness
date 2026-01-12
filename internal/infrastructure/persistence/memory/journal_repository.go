package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"mindfulness/internal/domain/journal"
)

// JournalRepository is an in-memory implementation for journal entries.
type JournalRepository struct {
	mu     sync.RWMutex
	byDate map[string]journal.Entry
}

func NewJournalRepository() *JournalRepository {
	return &JournalRepository{
		byDate: make(map[string]journal.Entry),
	}
}

func (r *JournalRepository) Save(_ context.Context, entry journal.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byDate[dateKey(entry.Date)] = entry
	return nil
}

func (r *JournalRepository) Latest(_ context.Context) (*journal.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.byDate) == 0 {
		return nil, journal.ErrNotFound
	}

	keys := make([]string, 0, len(r.byDate))
	for key := range r.byDate {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	latest := r.byDate[keys[len(keys)-1]]
	copy := latest
	return &copy, nil
}

func (r *JournalRepository) List(_ context.Context) ([]journal.Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.byDate) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(r.byDate))
	for key := range r.byDate {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]journal.Entry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, r.byDate[key])
	}
	return entries, nil
}

func dateKey(date time.Time) string {
	return date.UTC().Format("2006-01-02")
}
