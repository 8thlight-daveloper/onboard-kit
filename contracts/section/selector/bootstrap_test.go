package selector_test

import (
	"testing"

	"onboard-kit/internal/model"
)

func TestBootstrapSelectionReturnsHeldMarkerWithoutFiltering(t *testing.T) {
	t.Parallel()
	off := false
	input := model.Plan{Sections: []model.Section{{ID: "b", Bootstrap: false}, {ID: "a", Bootstrap: true}, {ID: "c", Bootstrap: true, Enabled: &off}}}
	got := newSelector().Select(input)
	if len(got) != 2 || got[0].ID != "a" || !got[0].Bootstrap || got[1].ID != "b" || got[1].Bootstrap {
		t.Fatalf("selector changed markers or used them as gates: %#v", got)
	}
	got[0].Bootstrap = false
	if !input.Sections[1].Bootstrap {
		t.Fatal("returned list aliases original section slots")
	}
}
