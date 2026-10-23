package loader_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresLoading(t *testing.T) {
	t.Parallel()
	testSectionFieldLoading(t, "requires.json")
}

func TestRequiresLoadResultsOwnTheirLists(t *testing.T) {
	t.Parallel()
	loader := newLoader()
	data := []byte("schemaVersion: 1\nid: access\nrequires: [b, a, b]\n")
	first, problems := loader.Load(data)
	if len(problems) != 0 || !reflect.DeepEqual(first.Requires, []string{"b", "a", "b"}) {
		t.Fatalf("requires not loaded: %#v, %#v", first, problems)
	}
	second, problems := loader.Load(data)
	if len(problems) != 0 || !reflect.DeepEqual(second.Requires, []string{"b", "a", "b"}) {
		t.Fatalf("second load: %#v, %#v", second, problems)
	}
	first.Requires[0] = "changed"
	if !reflect.DeepEqual(second.Requires, []string{"b", "a", "b"}) {
		t.Fatal("load results share requires")
	}
	for _, text := range []string{"requires: []", "requires: null", "", "requires: [a]"} {
		got, problems := loader.Load([]byte("schemaVersion: 1\nid: access\n" + text + "\n"))
		want := model.Section{ID: "access"}
		if text == "requires: [a]" {
			want.Requires = []string{"a"}
		}
		if len(problems) != 0 || !reflect.DeepEqual(got, want) {
			t.Fatalf("requires retained across calls: %#v, %#v; want %#v", got, problems, want)
		}
	}
}
