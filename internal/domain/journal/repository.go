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

// AdherenceRepository defines storage behavior for adherence state and logs.
// TODO: extract to `domain/adherence/repository.go`
type AdherenceRepository interface {
	Get(ctx context.Context) (Adherence, error)
	Save(ctx context.Context, adherence Adherence) error
	AppendLog(ctx context.Context, entry AdherenceLogEntry) error
}
