package adherence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

// Service coordinates adherence use cases.
type Service struct {
	repo journal.AdherenceRepository
	now  func() time.Time
}

func NewService(repo journal.AdherenceRepository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) Current(ctx context.Context) (journal.Adherence, error) {
	return s.repo.Get(ctx)
}

func (s *Service) Set(ctx context.Context, next journal.Adherence, notes map[journal.Precept]string) error {
	current, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	updated := make(journal.Adherence, len(current))
	for precept, value := range current {
		updated[precept] = value
	}

	for precept, value := range next {
		if !journal.IsKnownPrecept(precept) {
			return fmt.Errorf("unknown precept: %s", precept)
		}
		updated[precept] = value
	}

	if err := s.repo.Save(ctx, updated); err != nil {
		return err
	}

	now := s.now().UTC()
	for precept, from := range current {
		to := updated[precept]
		if from == to {
			continue
		}
		note := strings.TrimSpace(notes[precept])
		entry := journal.AdherenceLogEntry{
			At:      now,
			Precept: precept,
			From:    from,
			To:      to,
			Note:    note,
		}
		if err := s.repo.AppendLog(ctx, entry); err != nil {
			return err
		}
	}

	return nil
}
