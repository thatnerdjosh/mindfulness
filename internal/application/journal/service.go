package journal

import (
	"context"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

// Service coordinates journaling use cases.
type Service struct {
	repo journal.Repository
	now  func() time.Time
}

func NewService(repo journal.Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) RecordEntry(ctx context.Context, date time.Time, reflections map[journal.Precept]string, note string, mood string, foundation journal.Foundation) (journal.Entry, error) {
	entry, err := journal.NewEntry(date, reflections, note, mood, foundation, s.now())
	if err != nil {
		return journal.Entry{}, err
	}

	if err := s.repo.Save(ctx, entry); err != nil {
		return journal.Entry{}, err
	}

	return entry, nil
}

func (s *Service) LatestEntry(ctx context.Context) (*journal.Entry, error) {
	return s.repo.Latest(ctx)
}

func (s *Service) ListEntries(ctx context.Context) ([]journal.Entry, error) {
	return s.repo.List(ctx)
}
