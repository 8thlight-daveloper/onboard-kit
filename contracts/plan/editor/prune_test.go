package editor_test

import (
	"reflect"
	"slices"
	"testing"

	"onboard-kit/internal/model"
)

func TestPrunePreservesNilAndEmptyScopeMapsWhileRemovingOtherAnswers(t *testing.T) {
	t.Parallel()
	for _, withHidden := range []bool{false, true} {
		before := model.Plan{Answers: model.Answers{
			Sections: map[string]map[string]model.Value{
				"nil-scope":   nil,
				"empty-scope": {},
			},
		}}
		snapshot := model.Plan{Answers: model.Answers{
			Sections: map[string]map[string]model.Value{
				"nil-scope":   nil,
				"empty-scope": {},
			},
		}}
		if withHidden {
			before.Answers.Sections["gone"] = map[string]model.Value{
				"old": {Kind: model.ValueKindFlag, Flag: false},
			}
			snapshot.Answers.Sections["gone"] = map[string]model.Value{
				"old": {Kind: model.ValueKindFlag, Flag: false},
			}
		}
		got, cleared, problems := newEditor().PruneHiddenAnswers(before)
		if !reflect.DeepEqual(before, snapshot) {
			t.Fatal("prune changed the argument")
		}
		if answers, held := got.Answers.Sections["nil-scope"]; !held || answers != nil {
			t.Fatal("prune did not preserve a present nil scope map")
		}
		if answers, held := got.Answers.Sections["empty-scope"]; !held || answers == nil || len(answers) != 0 {
			t.Fatal("prune did not preserve a present empty scope map")
		}
		if _, held := got.Answers.Sections["gone"]; held {
			t.Fatal("prune retained the scope it emptied")
		}
		want := []string{}
		if withHidden {
			want = []string{"gone.old"}
		}
		if !slices.Equal(cleared.Answers, want) || len(cleared.Holders) != 0 || len(problems) != 0 {
			t.Fatalf("unexpected report: %+v %+v", cleared, problems)
		}
		if got.Answers.Plan != nil || got.Sections != nil {
			t.Fatal("prune filled absent editable fields")
		}
		got.Answers.Sections["new-scope"] = map[string]model.Value{}
		if !reflect.DeepEqual(before, snapshot) {
			t.Fatal("prune shares the outer scope map")
		}
	}
}
