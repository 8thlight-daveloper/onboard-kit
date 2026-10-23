package joiner_test

import (
	"testing"

	"onboard-kit/internal/model"
)

func TestBootstrapJoinPreservesDefinitionWhileLayeringAnswers(t *testing.T) {
	t.Parallel()
	input := model.Plan{Sections: []model.Section{{ID: "access", Bootstrap: true}, {ID: "other", Bootstrap: false}}, Answers: model.Answers{Plan: map[string]model.Value{"ready": {Kind: model.ValueKindFlag, Flag: false}}}}
	record := model.Record{Answers: model.Answers{Plan: map[string]model.Value{"ready": {Kind: model.ValueKindFlag, Flag: true}}}}
	got := newJoiner().Join(input, record)
	if len(got.Sections) != 2 || !got.Sections[0].Bootstrap || got.Sections[1].Bootstrap || !got.Answers.Plan["ready"].Flag {
		t.Fatalf("join dropped definition or did not layer record: %#v", got)
	}
	got.Sections[0].Bootstrap = false
	got.Sections[1].Bootstrap = true
	if !input.Sections[0].Bootstrap || input.Sections[1].Bootstrap || input.Answers.Plan["ready"].Flag {
		t.Fatal("joined values changed input")
	}
}
