package journal

import (
	"errors"
	"strings"
)

// Foundation represents the primary mindfulness foundation for the entry.
type Foundation string

const (
	FoundationKaya   Foundation = "kaya"
	FoundationVedana Foundation = "vedana"
	FoundationCit    Foundation = "cit"
	FoundationDhamma Foundation = "dhamma"
)

var ErrUnknownFoundation = errors.New("unknown foundation")

func IsKnownFoundation(foundation Foundation) bool {
	switch foundation {
	case FoundationKaya, FoundationVedana, FoundationCit, FoundationDhamma:
		return true
	default:
		return false
	}
}

func ParseFoundation(input string) (Foundation, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", "d", "dhamma":
		return FoundationDhamma, nil
	case "k", "kaya":
		return FoundationKaya, nil
	case "v", "vedana":
		return FoundationVedana, nil
	case "c", "cit":
		return FoundationCit, nil
	default:
		return "", ErrUnknownFoundation
	}
}

func FoundationLabel(foundation Foundation) string {
	switch foundation {
	case FoundationKaya:
		return "Kaya"
	case FoundationVedana:
		return "Vedana"
	case FoundationCit:
		return "Cit"
	case FoundationDhamma:
		return "Dhamma"
	default:
		return string(foundation)
	}
}
