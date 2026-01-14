package flatfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mindfulness/internal/domain/journal"
)

// DefaultAdherencePath returns the default JSON adherence file path.
func DefaultAdherencePath() (string, error) {
	dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "mt", "adherence.json"), nil
}

// DefaultAdherenceLogPath returns the default adherence log file path.
func DefaultAdherenceLogPath() (string, error) {
	dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "mt", "adherence.log.jsonl"), nil
}

// AdherenceRepository stores adherence state and log entries in flat files.
type AdherenceRepository struct {
	mu      sync.RWMutex
	path    string
	logPath string
	state   journal.Adherence
}

func NewAdherenceRepository(path string, logPath string) (*AdherenceRepository, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("adherence path is required")
	}
	logPath = strings.TrimSpace(logPath)
	if logPath == "" {
		return nil, fmt.Errorf("adherence log path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	repo := &AdherenceRepository{
		path:    path,
		logPath: logPath,
		state:   journal.DefaultAdherence(),
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *AdherenceRepository) Get(_ context.Context) (journal.Adherence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(journal.Adherence, len(r.state))
	for precept, value := range r.state {
		copy[precept] = value
	}
	return copy, nil
}

func (r *AdherenceRepository) Save(_ context.Context, adherence journal.Adherence) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := make(journal.Adherence, len(adherence))
	for precept, value := range adherence {
		copy[precept] = value
	}
	r.state = copy
	return r.persistLocked()
}

func (r *AdherenceRepository) AppendLog(_ context.Context, entry journal.AdherenceLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := adherenceLogRecord{
		Timestamp: entry.At.UTC().Format(time.RFC3339Nano),
		Precept:   string(entry.Precept),
		From:      entry.From,
		To:        entry.To,
		Note:      strings.TrimSpace(entry.Note),
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode adherence log entry: %w", err)
	}
	data = append(data, '\n')
	return appendFileAtomic(r.logPath, data, 0o600)
}

func (r *AdherenceRepository) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read adherence file: %w", err)
	}
	if len(data) == 0 {
		return nil
	}

	var record adherenceRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return fmt.Errorf("decode adherence file: %w", err)
	}

	state, err := record.toAdherence()
	if err != nil {
		return err
	}
	r.state = state
	return nil
}

func (r *AdherenceRepository) persistLocked() error {
	record := recordFromAdherence(r.state)
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode adherence file: %w", err)
	}
	data = append(data, '\n')
	return writeFileAtomic(r.path, data, 0o600)
}

type adherenceRecord map[string]bool

func recordFromAdherence(adherence journal.Adherence) adherenceRecord {
	precepts := make(map[string]bool, len(adherence))
	for precept, value := range adherence {
		precepts[string(precept)] = value
	}
	return adherenceRecord(precepts)
}

func (r adherenceRecord) toAdherence() (journal.Adherence, error) {
	state := journal.DefaultAdherence()
	for precept, value := range r {
		if !journal.IsKnownPrecept(journal.Precept(precept)) {
			return nil, fmt.Errorf("unknown precept in adherence file: %s", precept)
		}
		state[journal.Precept(precept)] = value
	}
	return state, nil
}

type adherenceLogRecord struct {
	Timestamp string `json:"timestamp"`
	Precept   string `json:"precept"`
	From      bool   `json:"from"`
	To        bool   `json:"to"`
	Note      string `json:"note,omitempty"`
}

func appendFileAtomic(path string, data []byte, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("append log file: %w", err)
	}
	return nil
}
