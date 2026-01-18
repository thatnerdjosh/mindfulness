package journal

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("journal entry not found")

// Repository defines storage behavior for journal entries.
type Repository interface {
	Save(ctx context.Context, entry Entry) error
	Latest(ctx context.Context) (*Entry, error)
	List(ctx context.Context) ([]Entry, error)
}
