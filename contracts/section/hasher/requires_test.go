package hasher_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestRequiresHashes(t *testing.T) {
	t.Parallel()
	var cases []struct {
		Name, Source, Canonical, Hash string
		Section                       json.RawMessage
	}
	if err := json.Unmarshal(readFixture(t, "requires.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 11 {
		t.Fatal("expected eleven independently pinned Requires hashes")
	}
	shared := newHasher()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Source == "" || !hashShape.MatchString(c.Hash) || textHash(c.Canonical) != c.Hash {
				t.Fatal("fixture source or independent canonical SHA-256 is wrong")
			}
			section, err := jsonform.Section(c.Section)
			if err != nil {
				t.Fatal(err)
			}
			before, err := jsonform.Section(c.Section)
			if err != nil {
				t.Fatal(err)
			}
			for call := 0; call < 3; call++ {
				if got := shared.Hash(section); got != c.Hash {
					t.Errorf("reused hasher gave %s; want %s for %s", got, c.Hash, c.Canonical)
				}
				if got := newHasher().Hash(section); got != c.Hash {
					t.Errorf("fresh hasher gave %s; want %s for %s", got, c.Hash, c.Canonical)
				}
				if !reflect.DeepEqual(section, before) {
					t.Fatal("Hash changed the Requires fixture")
				}
			}
		})
	}
}

func TestRequiresSetsDoNotMutateOverlappingInputs(t *testing.T) {
	t.Parallel()
	storage := []string{"z", "a", "m", "b"}
	section := model.Section{ID: "access", After: storage[:3], Requires: storage[1:]}
	before := []string{"z", "a", "m", "b"}
	canonical := `{"id":"access","timing":"anytime","after":["a","m","z"],"requires":["a","b","m"]}`
	const want = "bd8d30aa0aaddd58216c9844af822cc8490e4eb47bb1b4981cac8c04d0749659"
	if textHash(canonical) != want {
		t.Fatal("independent overlapping-list canonical hash is wrong")
	}
	hasher := newHasher()
	for call := 0; call < 3; call++ {
		if got := hasher.Hash(section); got != want {
			t.Errorf("got %s; want %s for independently held overlapping lists", got, want)
		}
		if !reflect.DeepEqual(storage, before) {
			t.Fatal("Hash sorted or deduplicated shared input storage in place")
		}
	}
}

func TestNilAndEmptyRequiresPreserveTheOldHash(t *testing.T) {
	t.Parallel()
	const want = "7f45f0901771cc2e49d31383a646ab9a5e0023a200a0156adc3cf480e3c708c9"
	for _, requires := range [][]string{nil, {}} {
		section := model.Section{Requires: requires}
		if got := newHasher().Hash(section); got != want {
			t.Errorf("Requires %#v: got %s; want old zero-section hash %s", requires, got, want)
		}
		if !reflect.DeepEqual(section.Requires, requires) {
			t.Fatal("Hash changed nil or empty list representation")
		}
	}
}
