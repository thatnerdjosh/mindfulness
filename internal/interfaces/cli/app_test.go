package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	adherenceapp "mindfulness/internal/application/adherence"
	journalapp "mindfulness/internal/application/journal"
	"mindfulness/internal/domain/journal"
	"mindfulness/internal/infrastructure/persistence/memory"
)

func TestRunJournalAddAndLatest(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournalAdd([]string{
		"--date=2024-01-02",
		"--note=steady day",
		"--mood=calm",
		"--reverence=grateful",
	}, svc, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "journaled 2024-01-02") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	out.Reset()
	err = runJournalLatest(svc, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "latest 2024-01-02") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalAddRequiresContent(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournalAdd([]string{
		"--date=2024-01-02",
	}, svc, &out, &errOut)
	if !errors.Is(err, journal.ErrEmptyEntry) {
		t.Fatalf("expected ErrEmptyEntry, got %v", err)
	}
}

func TestRunJournalListEmpty(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer

	if err := runJournalList(svc, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out.String()) != "no entries yet" {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalListWithEntries(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := runJournalAdd([]string{
		"--date=2024-01-01",
		"--note=steady",
	}, svc, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out.Reset()
	if err := runJournalAdd([]string{
		"--date=2024-01-03",
		"--note=focused",
	}, svc, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out.Reset()
	if err := runJournalList(svc, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "2024-01-01") || !strings.Contains(out.String(), "2024-01-03") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestParseDate(t *testing.T) {
	_, err := parseDate("invalid")
	if err == nil {
		t.Fatalf("expected error")
	}

	parsedDefault, err := parseDate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedDefault.IsZero() {
		t.Fatalf("expected default date to be set")
	}

	parsed, err := parseDate("2024-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Format("2006-01-02") != "2024-01-02" {
		t.Fatalf("unexpected date: %s", parsed.Format("2006-01-02"))
	}
}

func TestRunTopLevelHelpAndUnknown(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := Run([]string{"mt"}, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected usage output")
	}

	out.Reset()
	errOut.Reset()
	if err := Run([]string{"mt", "help"}, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "mt journal add") {
		t.Fatalf("expected help output")
	}

	out.Reset()
	errOut.Reset()
	if err := Run([]string{"mt", "version"}, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "mt") {
		t.Fatalf("expected version output")
	}

	out.Reset()
	errOut.Reset()
	if err := Run([]string{"mt", "unknown"}, &out, &errOut); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("expected unknown command output")
	}
}

func TestRunTopLevelJournalList(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := Run([]string{"mt", "journal", "list"}, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "no entries yet") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalHelpAndUnknown(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := runJournal([]string{"help"}, svc, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "journal add") {
		t.Fatalf("expected journal usage")
	}

	out.Reset()
	errOut.Reset()
	if err := runJournal([]string{"oops"}, svc, strings.NewReader(""), &out, &errOut); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(errOut.String(), "unknown journal command") {
		t.Fatalf("expected error output")
	}
}

func TestRunJournalRequiresSubcommand(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var errOut bytes.Buffer

	if err := runJournal([]string{}, svc, strings.NewReader(""), &bytes.Buffer{}, &errOut); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Fatalf("expected usage output")
	}
}

func TestRunJournalLatestWithNone(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)

	if err := runJournalLatest(svc, &bytes.Buffer{}); !errors.Is(err, journal.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRunJournalAddInvalidDate(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournalAdd([]string{
		"--date=bad-date",
		"--note=steady",
	}, svc, &out, &errOut)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunJournalAddFlagParseError(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournalAdd([]string{"--unknown"}, svc, &out, &errOut)
	if err == nil {
		t.Fatalf("expected error")
	}
}

type errorRepo struct{}

func (errorRepo) Save(_ context.Context, _ journal.Entry) error {
	return errors.New("save failed")
}

func (errorRepo) Latest(_ context.Context) (*journal.Entry, error) {
	return nil, errors.New("latest failed")
}

func (errorRepo) List(_ context.Context) ([]journal.Entry, error) {
	return nil, errors.New("list failed")
}

func TestRunJournalListError(t *testing.T) {
	svc := journalapp.NewService(errorRepo{})
	if err := runJournalList(svc, &bytes.Buffer{}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunJournalGuidedNoConfirm(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"2024-01-02",
		"calm",
		"steady note",
		"reverence reflection",
		"happiness reflection",
		"",
		"",
		"",
	}, "\n"))

	err := runJournalGuided([]string{"--no-confirm"}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "journaled 2024-01-02") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	latest, err := svc.LatestEntry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Date.Format("2006-01-02") != "2024-01-02" {
		t.Fatalf("unexpected date: %s", latest.Date.Format("2006-01-02"))
	}
}

func TestRunJournalGuidedConfirmNo(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"2024-01-03",
		"reflective",
		"note",
		"",
		"",
		"",
		"",
		"",
		"n",
	}, "\n"))

	err := runJournalGuided([]string{}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "not saved") {
		t.Fatalf("expected cancellation output, got %s", out.String())
	}

	list, err := svc.ListEntries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected no entries saved")
	}
}

func TestRunJournalGuidedRequiresContent(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"2024-01-04",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	}, "\n"))

	err := runJournalGuided([]string{"--no-confirm"}, svc, input, &out, &errOut)
	if !errors.Is(err, journal.ErrEmptyEntry) {
		t.Fatalf("expected ErrEmptyEntry, got %v", err)
	}
}

func TestRunJournalGuidedConfirmYes(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"2024-01-05",
		"grounded",
		"note",
		"",
		"reflection",
		"",
		"",
		"",
		"y",
	}, "\n"))

	err := runJournalGuided([]string{}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "journaled 2024-01-05") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalRoutesToGuided(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"2024-01-06",
		"",
		"note",
		"",
		"",
		"",
		"",
		"",
	}, "\n"))

	err := runJournal([]string{"guided", "--no-confirm"}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "journaled 2024-01-06") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalRoutesToAdd(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournal([]string{"add", "--date=2024-01-07", "--note=steady"}, svc, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "journaled 2024-01-07") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalRoutesToList(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournal([]string{"list"}, svc, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "no entries yet") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalRoutesToLatest(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := runJournalAdd([]string{"--date=2024-01-08", "--note=steady"}, svc, &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out.Reset()
	if err := runJournal([]string{"latest"}, svc, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "latest 2024-01-08") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestRunJournalGuidedInvalidDate(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"bad-date",
	}, "\n"))

	err := runJournalGuided([]string{"--no-confirm"}, svc, input, &out, &errOut)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunJournalGuidedFlagParseError(t *testing.T) {
	repo := memory.NewJournalRepository()
	svc := journalapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runJournalGuided([]string{"--unknown"}, svc, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatalf("expected error")
	}
}

type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestPromptReadError(t *testing.T) {
	reader := bufio.NewReader(errorReader{})
	_, err := prompt(reader, &bytes.Buffer{}, "prompt: ")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunAdherenceRequiresSubcommand(t *testing.T) {
	repo := memory.NewAdherenceRepository()
	svc := adherenceapp.NewService(repo)
	var errOut bytes.Buffer

	if err := runAdherence([]string{}, svc, strings.NewReader(""), &bytes.Buffer{}, &errOut); err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Fatalf("expected usage output")
	}
}

func TestRunAdherenceGuidedNoConfirm(t *testing.T) {
	repo := memory.NewAdherenceRepository()
	svc := adherenceapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"y",
		"n",
		"note for happiness",
		"",
		"",
		"",
	}, "\n"))

	err := runAdherenceGuided([]string{"--no-confirm"}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "adherence updated") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	state, err := svc.Current(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state[journal.TrueHappiness] {
		t.Fatalf("expected TrueHappiness false")
	}
}

func TestRunAdherenceGuidedConfirmNo(t *testing.T) {
	repo := memory.NewAdherenceRepository()
	svc := adherenceapp.NewService(repo)
	var out bytes.Buffer
	var errOut bytes.Buffer

	input := strings.NewReader(strings.Join([]string{
		"n",
		"note for reverence",
		"",
		"",
		"",
		"",
		"n",
	}, "\n"))

	err := runAdherenceGuided([]string{}, svc, input, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "not saved") {
		t.Fatalf("expected cancellation output, got %s", out.String())
	}

	state, err := svc.Current(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state[journal.ReverenceForLife] {
		t.Fatalf("expected ReverenceForLife to remain true")
	}
}
