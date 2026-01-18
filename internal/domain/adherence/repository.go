package adherence

import "context"

// Repository defines storage behavior for adherence state and logs.
type Repository interface {
	Get(ctx context.Context) (Adherence, error)
	Save(ctx context.Context, adherence Adherence) error
	AppendLog(ctx context.Context, entry AdherenceLogEntry) error
}
