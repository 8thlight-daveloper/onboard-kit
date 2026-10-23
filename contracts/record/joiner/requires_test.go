package joiner_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresJoinPreservesAndOwnsDefinition(t *testing.T) {
	t.Parallel()
	input := model.Plan{Sections: []model.Section{{ID: "access", Bootstrap: true, After: []string{"legacy"}, Requires: []string{"b", "a", "b"}}, {ID: "nil"}, {ID: "empty", Requires: []string{}}}, Answers: model.Answers{Plan: map[string]model.Value{"ready": {Kind: model.ValueKindFlag, Flag: false}}}}
	record := model.Record{Answers: model.Answers{Plan: map[string]model.Value{"ready": {Kind: model.ValueKindFlag, Flag: true}}}}
	got := newJoiner().Join(input, record)
	wantSections := []model.Section{{ID: "access", Bootstrap: true, After: []string{"legacy"}, Requires: []string{"b", "a", "b"}}, {ID: "nil"}, {ID: "empty", Requires: []string{}}}
	if !reflect.DeepEqual(got.Sections, wantSections) || !got.Answers.Plan["ready"].Flag {
		t.Fatalf("join changed definition or failed to layer answers: %#v", got)
	}
	if !reflect.DeepEqual(input.Sections, wantSections) || input.Answers.Plan["ready"].Flag {
		t.Fatal("input changed during join")
	}
	got.Sections[0].Requires[0] = "changed-result"
	if input.Sections[0].Requires[0] != "b" {
		t.Fatal("joined requires aliases input")
	}
	input.Sections[0].Requires[1] = "changed-input"
	if got.Sections[0].Requires[1] != "a" {
		t.Fatal("input requires aliases joined result")
	}
}
