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

	updated := make(adherence.Adherence, len(current))
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
