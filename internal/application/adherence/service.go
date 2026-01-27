package adherence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/adherence"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

// Service coordinates adherence use cases.
type Service struct {
	repo adherence.Repository
	now  func() time.Time
}

func NewService(repo adherence.Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) Current(ctx context.Context) (adherence.Adherence, error) {
	return s.repo.Get(ctx)
}

func (s *Service) Set(ctx context.Context, next adherence.Adherence, notes map[journal.Precept]string) error {
	current, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	updated, err := s.computeUpdatedAdherence(current, next)
	if err != nil {
		return err
	}

	if err := s.repo.Save(ctx, updated); err != nil {
		return err
	}

	return s.logChanges(ctx, current, updated, notes)
}

func (s *Service) computeUpdatedAdherence(current, next adherence.Adherence) (adherence.Adherence, error) {
	updated := make(adherence.Adherence, len(current))
	for precept, value := range current {
		updated[precept] = value
	}

	for precept, value := range next {
		if !journal.IsKnownPrecept(precept) {
			return nil, fmt.Errorf("unknown precept: %s", precept)
		}
		updated[precept] = value
	}

	return updated, nil
}

func (s *Service) logChanges(ctx context.Context, current, updated adherence.Adherence, notes map[journal.Precept]string) error {
	now := s.now().UTC()
	for precept, from := range current {
		to := updated[precept]
		if from == to {
			continue
		}
		note := strings.TrimSpace(notes[precept])
		entry := adherence.AdherenceLogEntry{
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
