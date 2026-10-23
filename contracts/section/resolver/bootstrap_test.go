package resolver_test

import (
	"testing"

	"onboard-kit/internal/model"
)

func TestBootstrapResolutionPreservesDefinition(t *testing.T) {
	t.Parallel()
	for _, marker := range []bool{false, true} {
		input := model.Section{ID: "access", Bootstrap: marker, Summary: "Use $account.", Variables: []model.Variable{{Name: "account"}}}
		answers := model.Answers{Sections: map[string]map[string]model.Value{"access": {"account": {Kind: model.ValueKindText, Text: "held"}}}}
		got, problems := newResolver().Resolve(input, answers, model.Directory{}, "")
		if len(problems) != 0 || got.Bootstrap != marker || got.Summary != "Use held." {
			t.Fatalf("resolved %#v with %#v", got, problems)
		}
		if input.Bootstrap != marker || input.Summary != "Use $account." {
			t.Fatal("input changed")
		}
	}
}
