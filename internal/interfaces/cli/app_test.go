package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	adherenceapp "github.com/thatnerdjosh/mindfulness/internal/application/adherence"
	journalapp "github.com/thatnerdjosh/mindfulness/internal/application/journal"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
	"github.com/thatnerdjosh/mindfulness/internal/infrastructure/persistence/memory"
)

func newInput(lines ...string) *strings.Reader {
	return strings.NewReader(strings.Join(lines, "\n"))
}

func TestRunJournalAdd(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantErr         error
		wantErrAny      bool
		wantOutContains string
	}{
		{
			name: "ok",
			args: []string{
				"--date=2024-01-02",
				"--note=steady day",
				"--mood=calm",
				"--reverence=grateful",
			},
			wantOutContains: "journaled 2024-01-02",
		},
		{
			name:    "requires content",
			args:    []string{"--date=2024-01-02"},
			wantErr: journal.ErrEmptyEntry,
		},
		{
			name:       "invalid date",
			args:       []string{"--date=bad-date", "--note=steady"},
			wantErrAny: true,
		},
		{
			name:       "flag parse error",
			args:       []string{"--unknown"},
			wantErrAny: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			var out bytes.Buffer
			var errOut bytes.Buffer

			err := runJournalAdd(tt.args, svc, &out, &errOut)
			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			case tt.wantErrAny:
				if err == nil {
					t.Fatalf("expected error")
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
					t.Fatalf("unexpected output: %s", out.String())
				}
			}
		})
	}
}

func TestRunJournalLatest(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(t *testing.T, svc *journalapp.Service)
		wantErr         error
		wantOutContains string
	}{
		{
			name:    "none",
			wantErr: journal.ErrNotFound,
		},
		{
			name: "with entry",
			setup: func(t *testing.T, svc *journalapp.Service) {
				var out bytes.Buffer
				var errOut bytes.Buffer
				err := runJournalAdd([]string{
					"--date=2024-01-02",
					"--note=steady day",
				}, svc, &out, &errOut)
				if err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}
			},
			wantOutContains: "latest 2024-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			if tt.setup != nil {
				tt.setup(t, svc)
			}

			var out bytes.Buffer
			err := runJournalLatest(svc, &out)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
				t.Fatalf("unexpected output: %s", out.String())
			}
		})
	}
}

func TestRunJournalList(t *testing.T) {
	tests := []struct {
		name            string
		buildService    func() *journalapp.Service
		setup           func(t *testing.T, svc *journalapp.Service)
		wantErrAny      bool
		wantOutContains []string
	}{
		{
			name:         "empty",
			buildService: func() *journalapp.Service { return journalapp.NewService(memory.NewJournalRepository()) },
			wantOutContains: []string{
				"no entries yet",
			},
		},
		{
			name:         "with entries",
			buildService: func() *journalapp.Service { return journalapp.NewService(memory.NewJournalRepository()) },
			setup: func(t *testing.T, svc *journalapp.Service) {
				var out bytes.Buffer
				var errOut bytes.Buffer
				if err := runJournalAdd([]string{"--date=2024-01-01", "--note=steady"}, svc, &out, &errOut); err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}
				out.Reset()
				if err := runJournalAdd([]string{"--date=2024-01-03", "--note=focused"}, svc, &out, &errOut); err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}
			},
			wantOutContains: []string{
				"2024-01-01",
				"2024-01-03",
			},
		},
		{
			name:         "error",
			buildService: func() *journalapp.Service { return journalapp.NewService(errorRepo{}) },
			wantErrAny:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.buildService()
			if tt.setup != nil {
				tt.setup(t, svc)
			}

			var out bytes.Buffer
			err := runJournalList(svc, &out)
			if tt.wantErrAny {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tt.wantOutContains {
				if !strings.Contains(out.String(), want) {
					t.Fatalf("unexpected output: %s", out.String())
				}
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantErr       bool
		wantZero      bool
		wantFormatted string
	}{
		{
			name:    "invalid",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:     "default",
			input:    "",
			wantZero: false,
		},
		{
			name:          "explicit",
			input:         "2024-01-02",
			wantFormatted: "2024-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantZero && !parsed.IsZero() {
				t.Fatalf("expected zero date")
			}
			if !tt.wantZero && parsed.IsZero() {
				t.Fatalf("expected default date to be set")
			}
			if tt.wantFormatted != "" && parsed.Format("2006-01-02") != tt.wantFormatted {
				t.Fatalf("unexpected date: %s", parsed.Format("2006-01-02"))
			}
		})
	}
}

func TestRunTopLevel(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantErr           bool
		wantOutContains   string
		wantErrOutContain string
	}{
		{
			name:            "usage",
			args:            []string{"mt"},
			wantOutContains: "Usage:",
		},
		{
			name:            "help",
			args:            []string{"mt", "help"},
			wantOutContains: "mt journal add",
		},
		{
			name:            "version",
			args:            []string{"mt", "version"},
			wantOutContains: "mt",
		},
		{
			name:              "unknown",
			args:              []string{"mt", "unknown"},
			wantErr:           true,
			wantErrOutContain: "unknown command",
		},
		{
			name:            "journal list",
			args:            []string{"mt", "journal", "list"},
			wantOutContains: "no entries yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			var out bytes.Buffer
			var errOut bytes.Buffer

			err := Run(tt.args, &out, &errOut)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tt.wantErrOutContain != "" && !strings.Contains(errOut.String(), tt.wantErrOutContain) {
					t.Fatalf("expected error output")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
				t.Fatalf("unexpected output: %s", out.String())
			}
		})
	}
}

func TestRunJournalCommands(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantErr           bool
		wantOutContains   string
		wantErrOutContain string
	}{
		{
			name:            "help",
			args:            []string{"help"},
			wantOutContains: "journal add",
		},
		{
			name:              "unknown",
			args:              []string{"oops"},
			wantErr:           true,
			wantErrOutContain: "unknown journal command",
		},
		{
			name:              "requires subcommand",
			args:              []string{},
			wantErr:           true,
			wantErrOutContain: "Usage:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			var out bytes.Buffer
			var errOut bytes.Buffer

			err := runJournal(tt.args, svc, strings.NewReader(""), &out, &errOut)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tt.wantErrOutContain != "" && !strings.Contains(errOut.String(), tt.wantErrOutContain) {
					t.Fatalf("expected error output")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
				t.Fatalf("unexpected output: %s", out.String())
			}
		})
	}
}

func TestRunJournalGuided(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		input           []string
		wantErr         error
		wantErrAny      bool
		wantOutContains string
		verify          func(t *testing.T, svc *journalapp.Service)
	}{
		{
			name: "no confirm",
			args: []string{"--no-confirm"},
			input: []string{
				"2024-01-02",
				"calm",
				"steady note",
				"reverence reflection",
				"happiness reflection",
				"",
				"",
				"",
			},
			wantOutContains: "journaled 2024-01-02",
			verify: func(t *testing.T, svc *journalapp.Service) {
				latest, err := svc.LatestEntry(context.Background())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if latest.Date.Format("2006-01-02") != "2024-01-02" {
					t.Fatalf("unexpected date: %s", latest.Date.Format("2006-01-02"))
				}
			},
		},
		{
			name: "confirm no",
			args: []string{},
			input: []string{
				"2024-01-03",
				"reflective",
				"note",
				"",
				"",
				"",
				"",
				"",
				"n",
			},
			wantOutContains: "not saved",
			verify: func(t *testing.T, svc *journalapp.Service) {
				list, err := svc.ListEntries(context.Background())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(list) != 0 {
					t.Fatalf("expected no entries saved")
				}
			},
		},
		{
			name: "requires content",
			args: []string{"--no-confirm"},
			input: []string{
				"2024-01-04",
				"",
				"",
				"",
				"",
				"",
				"",
				"",
			},
			wantErr: journal.ErrEmptyEntry,
		},
		{
			name: "confirm yes",
			args: []string{},
			input: []string{
				"2024-01-05",
				"grounded",
				"note",
				"",
				"reflection",
				"",
				"",
				"",
				"y",
			},
			wantOutContains: "journaled 2024-01-05",
		},
		{
			name:       "invalid date",
			args:       []string{"--no-confirm"},
			input:      []string{"bad-date"},
			wantErrAny: true,
		},
		{
			name:       "flag parse error",
			args:       []string{"--unknown"},
			wantErrAny: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			var out bytes.Buffer
			var errOut bytes.Buffer

			err := runJournalGuided(tt.args, svc, newInput(tt.input...), &out, &errOut)
			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			case tt.wantErrAny:
				if err == nil {
					t.Fatalf("expected error")
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
					t.Fatalf("unexpected output: %s", out.String())
				}
				if tt.verify != nil {
					tt.verify(t, svc)
				}
			}
		})
	}
}

func TestRunJournalRoutes(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		input           []string
		setup           func(t *testing.T, svc *journalapp.Service)
		wantOutContains string
	}{
		{
			name: "guided",
			args: []string{"guided", "--no-confirm"},
			input: []string{
				"2024-01-06",
				"",
				"note",
				"",
				"",
				"",
				"",
				"",
			},
			wantOutContains: "journaled 2024-01-06",
		},
		{
			name:            "add",
			args:            []string{"add", "--date=2024-01-07", "--note=steady"},
			wantOutContains: "journaled 2024-01-07",
		},
		{
			name:            "list",
			args:            []string{"list"},
			wantOutContains: "no entries yet",
		},
		{
			name: "latest",
			args: []string{"latest"},
			setup: func(t *testing.T, svc *journalapp.Service) {
				var out bytes.Buffer
				var errOut bytes.Buffer
				if err := runJournalAdd([]string{"--date=2024-01-08", "--note=steady"}, svc, &out, &errOut); err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}
			},
			wantOutContains: "latest 2024-01-08",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			if tt.setup != nil {
				tt.setup(t, svc)
			}

			var out bytes.Buffer
			var errOut bytes.Buffer
			err := runJournal(tt.args, svc, newInput(tt.input...), &out, &errOut)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
				t.Fatalf("unexpected output: %s", out.String())
			}
		})
	}
}

func TestPrompt(t *testing.T) {
	tests := []struct {
		name      string
		reader    func() *bufio.Reader
		wantErr   bool
		wantValue string
	}{
		{
			name:    "read error",
			reader:  func() *bufio.Reader { return bufio.NewReader(errorReader{}) },
			wantErr: true,
		},
		{
			name:      "trims input",
			reader:    func() *bufio.Reader { return bufio.NewReader(strings.NewReader("hello\n")) },
			wantValue: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := tt.reader()
			got, err := prompt(reader, &bytes.Buffer{}, "prompt: ")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantValue != "" && got != tt.wantValue {
				t.Fatalf("unexpected value: %s", got)
			}
		})
	}
}

func TestRunAdherence(t *testing.T) {
	type runFunc func(args []string, svc *adherenceapp.Service, input *strings.Reader, out, errOut *bytes.Buffer) error

	tests := []struct {
		name              string
		args              []string
		input             []string
		run               runFunc
		wantErr           bool
		wantOutContains   string
		wantErrOutContain string
		verify            func(t *testing.T, svc *adherenceapp.Service)
	}{
		{
			name: "requires subcommand",
			args: []string{},
			run: func(args []string, svc *adherenceapp.Service, input *strings.Reader, out, errOut *bytes.Buffer) error {
				return runAdherence(args, svc, input, out, errOut)
			},
			wantErr:           true,
			wantErrOutContain: "Usage:",
		},
		{
			name: "guided no confirm",
			args: []string{"--no-confirm"},
			input: []string{
				"y",
				"n",
				"note for happiness",
				"",
				"",
				"",
			},
			run: func(args []string, svc *adherenceapp.Service, input *strings.Reader, out, errOut *bytes.Buffer) error {
				return runAdherenceGuided(args, svc, input, out, errOut)
			},
			wantOutContains: "adherence updated",
			verify: func(t *testing.T, svc *adherenceapp.Service) {
				state, err := svc.Current(context.Background())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if state[journal.TrueHappiness] {
					t.Fatalf("expected TrueHappiness false")
				}
			},
		},
		{
			name: "guided confirm no",
			args: []string{},
			input: []string{
				"n",
				"note for reverence",
				"",
				"",
				"",
				"",
				"n",
			},
			run: func(args []string, svc *adherenceapp.Service, input *strings.Reader, out, errOut *bytes.Buffer) error {
				return runAdherenceGuided(args, svc, input, out, errOut)
			},
			wantOutContains: "not saved",
			verify: func(t *testing.T, svc *adherenceapp.Service) {
				state, err := svc.Current(context.Background())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !state[journal.ReverenceForLife] {
					t.Fatalf("expected ReverenceForLife to remain true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewAdherenceRepository()
			svc := adherenceapp.NewService(repo)
			var out bytes.Buffer
			var errOut bytes.Buffer

			err := tt.run(tt.args, svc, newInput(tt.input...), &out, &errOut)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tt.wantErrOutContain != "" && !strings.Contains(errOut.String(), tt.wantErrOutContain) {
					t.Fatalf("expected error output")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantOutContains != "" && !strings.Contains(out.String(), tt.wantOutContains) {
				t.Fatalf("unexpected output: %s", out.String())
			}
			if tt.verify != nil {
				tt.verify(t, svc)
			}
		})
	}
}

func TestRunQuicknote(t *testing.T) {
	tests := []struct {
		name            string
		input           []string
		wantOutContains string
	}{
		{
			name:            "adds note",
			input:           []string{"note"},
			wantOutContains: "journaled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewJournalRepository()
			svc := journalapp.NewService(repo)
			today, _ := parseDate("")

			var out bytes.Buffer
			var errOut bytes.Buffer

			err := runQuicknote([]string{}, svc, newInput(tt.input...), &out, &errOut)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(out.String(), tt.wantOutContains+" "+today.Format("2006-01-02")) {
				t.Fatalf("unexpected output: %s", out.String())
			}
		})
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

type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}
