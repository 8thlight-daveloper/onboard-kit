package encoder_test

import (
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresEncoding(t *testing.T) {
	t.Parallel()
	testSectionFieldEncoding(t, "requires.json")
}

func TestRequiresCollisionDoesNotInspectItsValue(t *testing.T) {
	t.Parallel()
	for _, requires := range [][]string{nil, {}, {"a"}} {
		input := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", Requires: requires, UnknownFields: map[string]any{"requires": make(chan int)}}}}
		data, problems := newEncoder().Encode(input)
		if len(data) != 0 || len(problems) != 1 || problems[0].Code != "encode.unknown-field-collision" || problems[0].Detail["path"] != "sections[0].requires" || problems[0].Severity != model.SeverityError {
			t.Fatalf("collision must precede unsupported-value traversal: %q, %#v", data, problems)
		}
	}
}

func TestRequiresEncoderStateAcrossCalls(t *testing.T) {
	t.Parallel()
	encoder := newEncoder()
	for _, c := range []struct {
		requires []string
		output   string
	}{{[]string{"b", "a", "b"}, "    requires:\n      - b\n      - a\n      - b\n"}, {nil, ""}, {[]string{}, ""}, {[]string{"a"}, "    requires:\n      - a\n"}} {
		input := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", Requires: c.requires}}}
		got, problems := encoder.Encode(input)
		want := "schemaVersion: 1\nsections:\n  - id: access\n" + c.output
		if len(problems) != 0 || string(got) != want {
			t.Fatalf("requires retained across calls: %q, %#v; want %q", got, problems, want)
		}
	}
}
