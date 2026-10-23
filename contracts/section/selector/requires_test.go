package selector_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresSelectionPreservesHeldListsInReferenceOrder(t *testing.T) {
	t.Parallel()
	input := model.Plan{Sections: []model.Section{{ID: "c"}, {ID: "a", Bootstrap: true, Requires: []string{"c", "b", "c"}}, {ID: "b"}}}
	want := []model.Section{{ID: "a", Bootstrap: true, Requires: []string{"c", "b", "c"}}, {ID: "b"}, {ID: "c"}}
	got := newSelector().Select(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selection changed requires or used it as reference ordering: %#v; want %#v", got, want)
	}
	if input.Sections[0].ID != "c" || !reflect.DeepEqual(input.Sections[1].Requires, []string{"c", "b", "c"}) {
		t.Fatal("selection changed input")
	}
	got[0].Requires = []string{"replacement"}
	if !reflect.DeepEqual(input.Sections[1].Requires, []string{"c", "b", "c"}) {
		t.Fatal("result section slots alias input")
	}
}
