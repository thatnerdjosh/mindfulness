package journal

import "testing"

func TestAllPrecepts(t *testing.T) {
	precepts := AllPrecepts()
	if len(precepts) != 5 {
		t.Fatalf("expected 5 precepts, got %d", len(precepts))
	}

	for _, info := range precepts {
		if info.ID == "" {
			t.Fatalf("expected precept id to be set")
		}
		if info.Title == "" {
			t.Fatalf("expected precept title to be set")
		}
		if !IsKnownPrecept(info.ID) {
			t.Fatalf("expected precept to be known: %s", info.ID)
		}
	}

	if IsKnownPrecept(Precept("unknown")) {
		t.Fatalf("expected unknown precept to be false")
	}
}
