package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	adherenceapp "github.com/thatnerdjosh/mindfulness/internal/application/adherence"
	journalapp "github.com/thatnerdjosh/mindfulness/internal/application/journal"
	"github.com/thatnerdjosh/mindfulness/internal/domain/journal"
	"github.com/thatnerdjosh/mindfulness/internal/infrastructure/persistence/flatfile"
)

const version = "0.1.0"

// Run executes the CLI application.
func Run(args []string, out io.Writer, errOut io.Writer) error {
	if len(args) < 2 {
		printUsage(out)
		return nil
	}

	repoPath, err := flatfile.DefaultJournalPath()
	if err != nil {
		return err
	}
	repo, err := flatfile.NewJournalRepository(repoPath)
	if err != nil {
		return err
	}
	svc := journalapp.NewService(repo)

	adherencePath, err := flatfile.DefaultAdherencePath()
	if err != nil {
		return err
	}
	adherenceLogPath, err := flatfile.DefaultAdherenceLogPath()
	if err != nil {
		return err
	}
	adherenceRepo, err := flatfile.NewAdherenceRepository(adherencePath, adherenceLogPath)
	if err != nil {
		return err
	}
	adherenceSvc := adherenceapp.NewService(adherenceRepo)

	switch args[1] {
	case "version", "-v", "--version":
		fmt.Fprintln(out, "mt", version)
		return nil
	case "journal":
		return runJournal(args[2:], svc, os.Stdin, out, errOut)
	case "quicknote":
		return runQuicknote(args[2:], svc, os.Stdin, out, errOut)
	case "adherence":
		return runAdherence(args[2:], adherenceSvc, os.Stdin, out, errOut)
	case "help", "-h", "--help":
		printUsage(out)
		return nil
	default:
		fmt.Fprintf(errOut, "unknown command: %s\n", args[1])
		printUsage(errOut)
		return fmt.Errorf("unknown command: %s", args[1])
	}
}

func runAdherence(args []string, svc *adherenceapp.Service, in io.Reader, out io.Writer, errOut io.Writer) error {
	if len(args) < 1 {
		printAdherenceUsage(errOut)
		return fmt.Errorf("adherence subcommand required")
	}

	switch args[0] {
	case "guided":
		return runAdherenceGuided(args[1:], svc, in, out, errOut)
	case "help", "-h", "--help":
		printAdherenceUsage(out)
		return nil
	default:
		fmt.Fprintf(errOut, "unknown adherence command: %s\n", args[0])
		printAdherenceUsage(errOut)
		return fmt.Errorf("unknown adherence command: %s", args[0])
	}
}

func runJournal(args []string, svc *journalapp.Service, in io.Reader, out io.Writer, errOut io.Writer) error {
	if len(args) < 1 {
		printJournalUsage(errOut)
		return fmt.Errorf("journal subcommand required")
	}

	switch args[0] {
	case "add":
		return runJournalAdd(args[1:], svc, out, errOut)
	case "guided":
		return runJournalGuided(args[1:], svc, in, out, errOut)
	case "latest":
		return runJournalLatest(svc, out)
	case "list":
		return runJournalList(svc, out)
	case "help", "-h", "--help":
		printJournalUsage(out)
		return nil
	default:
		fmt.Fprintf(errOut, "unknown journal command: %s\n", args[0])
		printJournalUsage(errOut)
		return fmt.Errorf("unknown journal command: %s", args[0])
	}
}

func runQuicknote(args []string, svc *journalapp.Service, in io.Reader, out io.Writer, errOut io.Writer) error {
	reader := bufio.NewReader(in)
	note, err := prompt(reader, out, "Quicknote: ")
	if note == "" {
		// TODO: Extract to const for error string
		return errors.New("Unable to create quicknote without content.")
	}

	if err != nil {
		return err
	}

	date, err := parseDate("")
	if err != nil {
		return err
	}

	entry, err := svc.RecordEntry(context.Background(), date, map[journal.Precept]string{}, note, "")
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "journaled %s reflections=%d mood=%s\n", entry.Date.Format("2006-01-02"), len(entry.Reflections), entry.Mood)
	return nil
}

func runJournalAdd(args []string, svc *journalapp.Service, out io.Writer, errOut io.Writer) error {
	fs := flag.NewFlagSet("journal add", flag.ContinueOnError)
	fs.SetOutput(errOut)
	dateStr := fs.String("date", "", "date in YYYY-MM-DD (defaults to today)")
	note := fs.String("note", "", "overall note")
	mood := fs.String("mood", "", "overall mood")
	reverence := fs.String("reverence", "", "reflection on Reverence For Life")
	happiness := fs.String("happiness", "", "reflection on True Happiness")
	love := fs.String("love", "", "reflection on True Love")
	speech := fs.String("speech", "", "reflection on Loving Speech and Deep Listening")
	nourishment := fs.String("nourishment", "", "reflection on Nourishment and Healing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	date, err := parseDate(*dateStr)
	if err != nil {
		return err
	}

	reflections := map[journal.Precept]string{
		journal.ReverenceForLife:          *reverence,
		journal.TrueHappiness:             *happiness,
		journal.TrueLove:                  *love,
		journal.LovingSpeechDeepListening: *speech,
		journal.NourishmentAndHealing:     *nourishment,
	}

	entry, err := svc.RecordEntry(context.Background(), date, reflections, *note, *mood)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "journaled %s reflections=%d mood=%s\n", entry.Date.Format("2006-01-02"), len(entry.Reflections), entry.Mood)
	return nil
}

func runJournalGuided(args []string, svc *journalapp.Service, in io.Reader, out io.Writer, errOut io.Writer) error {
	fs := flag.NewFlagSet("journal guided", flag.ContinueOnError)
	fs.SetOutput(errOut)
	noConfirm := fs.Bool("no-confirm", false, "save without confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	reader := bufio.NewReader(in)
	dateInput, err := prompt(reader, out, "Date (YYYY-MM-DD, default today): ")
	if err != nil {
		return err
	}
	date, err := parseDate(dateInput)
	if err != nil {
		return err
	}

	mood, err := prompt(reader, out, "Mood (optional): ")
	if err != nil {
		return err
	}
	note, err := prompt(reader, out, "Overall note (optional): ")
	if err != nil {
		return err
	}

	reflections := make(map[journal.Precept]string)
	for _, info := range journal.AllPrecepts() {
		question := fmt.Sprintf("%s reflection (optional): ", info.Title)
		reflection, err := prompt(reader, out, question)
		if err != nil {
			return err
		}
		reflection = strings.TrimSpace(reflection)
		if reflection != "" {
			reflections[info.ID] = reflection
		}
	}

	if strings.TrimSpace(note) == "" && len(reflections) == 0 {
		return journal.ErrEmptyEntry
	}

	if !*noConfirm {
		printGuidedSummary(out, date, mood, note, reflections)
		confirm, err := prompt(reader, out, "Save? (y/n): ")
		if err != nil {
			return err
		}
		if !isYes(confirm) {
			fmt.Fprintln(out, "not saved")
			return nil
		}
	}

	entry, err := svc.RecordEntry(context.Background(), date, reflections, note, mood)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "journaled %s reflections=%d mood=%s\n", entry.Date.Format("2006-01-02"), len(entry.Reflections), entry.Mood)
	return nil
}

func runJournalLatest(svc *journalapp.Service, out io.Writer) error {
	entry, err := svc.LatestEntry(context.Background())
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "latest %s reflections=%d mood=%s\n", entry.Date.Format("2006-01-02"), len(entry.Reflections), entry.Mood)
	return nil
}

func runJournalList(svc *journalapp.Service, out io.Writer) error {
	entries, err := svc.ListEntries(context.Background())
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(out, "no entries yet")
		return nil
	}

	for _, entry := range entries {
		fmt.Fprintf(out, "%s reflections=%d mood=%s\n", entry.Date.Format("2006-01-02"), len(entry.Reflections), entry.Mood)
	}
	return nil
}

func runAdherenceGuided(args []string, svc *adherenceapp.Service, in io.Reader, out io.Writer, errOut io.Writer) error {
	fs := flag.NewFlagSet("adherence guided", flag.ContinueOnError)
	fs.SetOutput(errOut)
	noConfirm := fs.Bool("no-confirm", false, "save without confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	current, err := svc.Current(context.Background())
	if err != nil {
		return err
	}

	reader := bufio.NewReader(in)
	next := make(journal.Adherence, len(current))
	notes := make(map[journal.Precept]string)

	for _, info := range journal.AllPrecepts() {
		currentValue := current[info.ID]
		question := fmt.Sprintf("%s (currently %s) keep? (y/n, default %s): ",
			info.Title,
			yesNoLabel(currentValue),
			yesNoLabel(currentValue),
		)
		answer, err := prompt(reader, out, question)
		if err != nil {
			return err
		}
		value, err := parseYesNoDefault(answer, currentValue)
		if err != nil {
			return err
		}
		next[info.ID] = value

		if value != currentValue {
			note, err := prompt(reader, out, fmt.Sprintf("Note for %s (optional): ", info.Title))
			if err != nil {
				return err
			}
			note = strings.TrimSpace(note)
			if note != "" {
				notes[info.ID] = note
			}
		}
	}

	if !*noConfirm {
		printAdherenceSummary(out, current, next, notes)
		confirm, err := prompt(reader, out, "Save? (y/n): ")
		if err != nil {
			return err
		}
		if !isYes(confirm) {
			fmt.Fprintln(out, "not saved")
			return nil
		}
	}

	if err := svc.Set(context.Background(), next, notes); err != nil {
		return err
	}

	fmt.Fprintln(out, "adherence updated")
	return nil
}

func prompt(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	fmt.Fprint(out, label)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func printGuidedSummary(out io.Writer, date time.Time, mood string, note string, reflections map[journal.Precept]string) {
	fmt.Fprintln(out, "Summary:")
	fmt.Fprintf(out, "Date: %s\n", date.Format("2006-01-02"))
	if strings.TrimSpace(mood) != "" {
		fmt.Fprintf(out, "Mood: %s\n", strings.TrimSpace(mood))
	}
	if strings.TrimSpace(note) != "" {
		fmt.Fprintf(out, "Note: %s\n", strings.TrimSpace(note))
	}
	for _, info := range journal.AllPrecepts() {
		if reflection, ok := reflections[info.ID]; ok {
			fmt.Fprintf(out, "%s: %s\n", info.Title, reflection)
		}
	}
}

func printAdherenceSummary(out io.Writer, current journal.Adherence, next journal.Adherence, notes map[journal.Precept]string) {
	fmt.Fprintln(out, "Summary:")
	for _, info := range journal.AllPrecepts() {
		before := current[info.ID]
		after := next[info.ID]
		if before == after {
			continue
		}
		fmt.Fprintf(out, "%s: %s -> %s\n", info.Title, yesNoLabel(before), yesNoLabel(after))
		if note, ok := notes[info.ID]; ok && strings.TrimSpace(note) != "" {
			fmt.Fprintf(out, "Note: %s\n", strings.TrimSpace(note))
		}
	}
}

func isYes(input string) bool {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func parseYesNoDefault(input string, defaultValue bool) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "":
		return defaultValue, nil
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("enter y or n")
	}
}

func yesNoLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func parseDate(input string) (time.Time, error) {
	if strings.TrimSpace(input) == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	parsed, err := time.Parse("2006-01-02", input)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %w", err)
	}
	return parsed.UTC(), nil
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "mindfulness (mt) - daily precept journal")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  mt journal add --date=YYYY-MM-DD --note=\"...\" --mood=... \\")
	fmt.Fprintln(out, "    --reverence=\"...\" --happiness=\"...\" --love=\"...\" --speech=\"...\" --nourishment=\"...\"")
	fmt.Fprintln(out, "  mt journal guided")
	fmt.Fprintln(out, "  mt journal latest")
	fmt.Fprintln(out, "  mt journal list")
	fmt.Fprintln(out, "  mt quicknote")
	fmt.Fprintln(out, "  mt adherence guided")
	fmt.Fprintln(out, "  mt version")
}

func printJournalUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  mt journal add --date=YYYY-MM-DD --note=\"...\" --mood=... \\")
	fmt.Fprintln(out, "    --reverence=\"...\" --happiness=\"...\" --love=\"...\" --speech=\"...\" --nourishment=\"...\"")
	fmt.Fprintln(out, "  mt journal guided")
	fmt.Fprintln(out, "  mt journal latest")
	fmt.Fprintln(out, "  mt journal list")
}

func printAdherenceUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  mt adherence guided")
}
