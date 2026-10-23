package instantiator_test

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresInstantiationPreservesAndOwnsDefinition(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name            string
		template        model.Section
		canonical, hash string
	}{
		{"list-with-repeats", model.Section{ID: "access", Requires: []string{"b", "a", "b"}}, `{"id":"access","timing":"anytime","requires":["a","b"]}`, "4f6a5ba7ddefab56aa399697f6e2f03f69aef6939f996aa1b1f1bc0c30ad0119"},
		{"independent-after-and-bootstrap", model.Section{ID: "access", Parent: "group", Bootstrap: true, Timing: model.Timing("week-one"), After: []string{"z", "m", "z"}, Requires: []string{"b", "a", "b"}, Owner: "lead"}, `{"id":"access","parent":"group","bootstrap":true,"timing":"week-one","after":["m","z"],"requires":["a","b"],"owner":"lead"}`, "20698d2da6855249dcf8bc4b91642ae83bc066434df83d3e16c7222158d4b367"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if fmt.Sprintf("%x", sha256.Sum256([]byte(c.canonical))) != c.hash {
				t.Fatal("handwritten canonical fixture and pinned hash disagree")
			}
			template := c.template
			template.Requires = append([]string(nil), c.template.Requires...)
			got, problems := newInstantiator().Instantiate([]model.Section{template}, nil, "set", "2026-01-02")
			want := c.template
			want.Source = &model.Source{Template: "access", Hash: c.hash}
			if len(problems) != 0 || len(got.Sections) != 1 || !reflect.DeepEqual(got.Sections[0], want) {
				t.Fatalf("section %#v, problems %#v; want %#v", got.Sections, problems, want)
			}
			if !reflect.DeepEqual(template, c.template) {
				t.Fatal("template changed while instantiating")
			}
			got.Sections[0].Requires[0] = "changed-result"
			if !reflect.DeepEqual(template.Requires, []string{"b", "a", "b"}) {
				t.Fatal("returned requires aliases template")
			}
			template.Requires[1] = "changed-input"
			if got.Sections[0].Requires[1] != "a" {
				t.Fatal("template requires aliases result")
			}
		})
	}
}
