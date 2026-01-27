package journal

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidDate    = errors.New("date is required")
	ErrEmptyEntry     = errors.New("entry must include a reflection or note")
	ErrUnknownPrecept = errors.New("unknown precept")
)

// Entry captures a daily mindfulness reflection.
type Entry struct {
	Date        time.Time
	Timestamp   time.Time
	Reflections map[Precept]string
	Note        string
	Mood        string
	Foundation  Foundation
}

func NewEntry(date time.Time, reflections map[Precept]string, note string, mood string, foundation Foundation, timestamp time.Time) (Entry, error) {
	if date.IsZero() {
		return Entry{}, ErrInvalidDate
	}

	cleanedReflections, err := validateAndCleanReflections(reflections)
	if err != nil {
		return Entry{}, err
	}

	note, mood = strings.TrimSpace(note), strings.TrimSpace(mood)
	if len(cleanedReflections) == 0 && note == "" {
		return Entry{}, ErrEmptyEntry
	}

	foundation, err = validateFoundation(foundation)
	if err != nil {
		return Entry{}, err
	}

	if timestamp.IsZero() {
		timestamp = normalizeDate(date)
	}

	return Entry{
		Date:        normalizeDate(date),
		Timestamp:   timestamp,
		Reflections: cleanedReflections,
		Note:        note,
		Mood:        mood,
		Foundation:  foundation,
	}, nil
}

func validateAndCleanReflections(reflections map[Precept]string) (map[Precept]string, error) {
	cleaned := make(map[Precept]string)
	for precept, reflection := range reflections {
		if !IsKnownPrecept(precept) {
			return nil, ErrUnknownPrecept
		}
		reflection = strings.TrimSpace(reflection)
		if reflection != "" {
			cleaned[precept] = reflection
		}
	}
	return cleaned, nil
}

func validateFoundation(foundation Foundation) (Foundation, error) {
	if foundation == "" {
		return FoundationDhamma, nil
	}
	if !IsKnownFoundation(foundation) {
		return "", ErrUnknownFoundation
	}
	return foundation, nil
}

func (e Entry) SortedPrecepts() []Precept {
	list := make([]Precept, 0, len(e.Reflections))
	for precept := range e.Reflections {
		list = append(list, precept)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})
	return list
}

func normalizeDate(date time.Time) time.Time {
	utc := date.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
