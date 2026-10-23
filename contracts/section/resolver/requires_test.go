package resolver_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresResolutionPreservesAndOwnsIdentifiers(t *testing.T) {
	t.Parallel()
	input := model.Section{ID: "access", Bootstrap: true, Summary: "Use $account.", After: []string{"legacy"}, Requires: []string{"$account", "b", "$account"}, Variables: []model.Variable{{Name: "account"}}}
	answers := model.Answers{Sections: map[string]map[string]model.Value{"access": {"account": {Kind: model.ValueKindText, Text: "held"}}}}
	got, problems := newResolver().Resolve(input, answers, model.Directory{}, "")
	want := model.Section{ID: "access", Bootstrap: true, Summary: "Use held.", After: []string{"legacy"}, Requires: []string{"$account", "b", "$account"}, Variables: []model.Variable{{Name: "account"}}}
	if len(problems) != 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved %#v, problems %#v; want %#v", got, problems, want)
	}
	if input.Summary != "Use $account." || !reflect.DeepEqual(input.Requires, []string{"$account", "b", "$account"}) {
		t.Fatal("input changed during resolution")
	}
	got.Requires[0] = "changed-result"
	if input.Requires[0] != "$account" {
		t.Fatal("result requires aliases input")
	}
	input.Requires[1] = "changed-input"
	if got.Requires[1] != "b" {
		t.Fatal("input requires aliases result")
	}
}
