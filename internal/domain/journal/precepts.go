package journal

// Precept identifies a mindfulness training.
type Precept string

const (
	ReverenceForLife          Precept = "reverence-for-life"
	TrueHappiness             Precept = "true-happiness"
	TrueLove                  Precept = "true-love"
	LovingSpeechDeepListening Precept = "loving-speech-deep-listening"
	NourishmentAndHealing     Precept = "nourishment-and-healing"
)

type PreceptInfo struct {
	ID    Precept
	Title string
}

func AllPrecepts() []PreceptInfo {
	return []PreceptInfo{
		{ID: ReverenceForLife, Title: "Reverence For Life"},
		{ID: TrueHappiness, Title: "True Happiness"},
		{ID: TrueLove, Title: "True Love"},
		{ID: LovingSpeechDeepListening, Title: "Loving Speech and Deep Listening"},
		{ID: NourishmentAndHealing, Title: "Nourishment and Healing"},
	}
}

func IsKnownPrecept(p Precept) bool {
	for _, info := range AllPrecepts() {
		if info.ID == p {
			return true
		}
	}
	return false
}
