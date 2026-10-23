package hasher_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
)

func TestBootstrapHashes(t *testing.T) {
	t.Parallel()
	var cases []struct {
		Name, Source, Canonical, Hash string
		Section                       json.RawMessage
	}
	if err := json.Unmarshal(readFixture(t, "bootstrap.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 4 {
		t.Fatal("expected four independently pinned bootstrap hashes")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if c.Source == "" || textHash(c.Canonical) != c.Hash {
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
			hasher := newHasher()
			if got := hasher.Hash(section); got != c.Hash {
				t.Fatalf("hash %s; want %s for %s", got, c.Hash, c.Canonical)
			}
			if !reflect.DeepEqual(section, before) || hasher.Hash(section) != c.Hash {
				t.Fatal("input changed or hash was nondeterministic")
			}
			section.Bootstrap = !section.Bootstrap
			if hasher.Hash(section) == c.Hash {
				t.Fatal("changing bootstrap did not change the definition hash")
			}
		})
	}
}
