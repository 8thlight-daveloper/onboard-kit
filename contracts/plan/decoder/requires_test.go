package decoder_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresDecoding(t *testing.T) {
	t.Parallel()
	testSectionFieldDecoding(t, "requires.json")
}

func TestRequiresDecodeResultsOwnTheirLists(t *testing.T) {
	t.Parallel()
	decoder := newDecoder()
	data := []byte("schemaVersion: 1\nsections:\n  - id: access\n    requires: [b, a, b]\n")
	first, problems := decoder.Decode(data)
	if len(problems) != 0 || len(first.Sections) != 1 || !reflect.DeepEqual(first.Sections[0].Requires, []string{"b", "a", "b"}) {
		t.Fatalf("requires not decoded: %#v, %#v", first, problems)
	}
	second, problems := decoder.Decode(data)
	if len(problems) != 0 || len(second.Sections) != 1 || !reflect.DeepEqual(second.Sections[0].Requires, []string{"b", "a", "b"}) {
		t.Fatalf("second decode: %#v, %#v", second, problems)
	}
	first.Sections[0].Requires[0] = "changed"
	if !reflect.DeepEqual(second.Sections[0].Requires, []string{"b", "a", "b"}) {
		t.Fatal("decode results share requires")
	}
	for _, text := range []string{"requires: []", "requires: null", "", "requires: [a]"} {
		got, problems := decoder.Decode([]byte("schemaVersion: 1\nsections:\n  - id: access\n    " + text + "\n"))
		want := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access"}}}
		if text == "requires: [a]" {
			want.Sections[0].Requires = []string{"a"}
		}
		if len(problems) != 0 || !reflect.DeepEqual(got, want) {
			t.Fatalf("requires retained across calls: %#v, %#v; want %#v", got, problems, want)
		}
	}
}
