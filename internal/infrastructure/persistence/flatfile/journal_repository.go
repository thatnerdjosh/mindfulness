package flatfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
)

// DefaultJournalPath returns the default JSON journal file path.
func DefaultJournalPath() (string, error) {
	dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "mt", "journal.json"), nil
}

// JournalRepository stores journal entries in a JSON file.
type JournalRepository struct {
	mu      sync.RWMutex
	path    string
	entries []journal.Entry
}

func NewJournalRepository(path string) (*JournalRepository, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("journal path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	repo := &JournalRepository{
		path:    path,
		entries: []journal.Entry{},
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JournalRepository) Save(_ context.Context, entry journal.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = append(r.entries, entry)
	return r.persistLocked()
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

func (r *JournalRepository) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read journal file: %w", err)
	}
	if len(data) == 0 {
		return nil
	}

	var records []entryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode journal file: %w", err)
	}

	for _, record := range records {
		entry, err := record.toEntry()
		if err != nil {
			return err
		}
		r.entries = append(r.entries, entry)
	}
	return nil
}

func (r *JournalRepository) persistLocked() error {
	entries := append([]journal.Entry{}, r.entries...)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})
	records := make([]entryRecord, 0, len(entries))
	for _, entry := range entries {
		records = append(records, recordFromEntry(entry))
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("encode journal file: %w", err)
	}
	data = append(data, '\n')
	return writeFileAtomic(r.path, data, 0o600)
}

type entryRecord struct {
	Date        string            `json:"date"`
	Reflections map[string]string `json:"reflections,omitempty"`
	Note        string            `json:"note,omitempty"`
	Mood        string            `json:"mood,omitempty"`
	Foundation  string            `json:"foundation,omitempty"`
}

func recordFromEntry(entry journal.Entry) entryRecord {
	reflections := make(map[string]string, len(entry.Reflections))
	for precept, reflection := range entry.Reflections {
		reflections[string(precept)] = reflection
	}
	return entryRecord{
		Date:        entry.Date.UTC().Format("2006-01-02"),
		Reflections: reflections,
		Note:        entry.Note,
		Mood:        entry.Mood,
		Foundation:  string(entry.Foundation),
	}
}

func (r entryRecord) toEntry() (journal.Entry, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(r.Date))
	if err != nil {
		return journal.Entry{}, fmt.Errorf("invalid journal date %q: %w", r.Date, err)
	}

	reflections := make(map[journal.Precept]string, len(r.Reflections))
	for precept, reflection := range r.Reflections {
		reflections[journal.Precept(precept)] = reflection
	}

	entry, err := journal.NewEntry(parsed, reflections, r.Note, r.Mood, journal.Foundation(strings.ToLower(strings.TrimSpace(r.Foundation))))
	if err != nil {
		return journal.Entry{}, fmt.Errorf("invalid journal entry for %s: %w", r.Date, err)
	}
	return entry, nil
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".journal-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace journal file: %w", err)
	}
	return nil
}
