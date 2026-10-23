package encoder_test

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func TestAllKnownSectionKeysRejectUnknownFieldCollisions(t *testing.T) {
	t.Parallel()
	keys := []string{
		"id", "title", "summary", "includeIf", "parent", "bootstrap", "timing",
		"enabled", "after", "owner", "escalation", "verification", "body",
		"variables", "steps", "source",
	}
	for _, key := range keys {
		for _, value := range []struct {
			name string
			data any
		}{{"unsupported", make(chan int)}, {"scalar", "held"}} {
			t.Run(key+"/"+value.name, func(t *testing.T) {
				t.Parallel()
				input := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", UnknownFields: map[string]any{key: value.data}}}}
				data, problems := newEncoder().Encode(input)
				if len(data) != 0 || len(problems) != 1 {
					t.Fatalf("expected no bytes and one known-key collision; got %q, %#v", data, problems)
				}
				problem := problems[0]
				if problem.Code != "encode.unknown-field-collision" || problem.Detail["path"] != "sections[0]."+key || problem.Severity != model.SeverityError || strings.TrimSpace(problem.Message) == "" {
					t.Fatalf("expected collision before value traversal at sections[0].%s; got %#v", key, problem)
				}
				if len(input.Sections[0].UnknownFields) != 1 || input.Sections[0].UnknownFields[key] != value.data {
					t.Fatal("colliding input changed")
				}
			})
		}
	}
}
