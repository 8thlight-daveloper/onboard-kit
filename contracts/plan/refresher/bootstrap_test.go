package refresher_test

import (
	"testing"

	"onboard-kit/internal/model"
)

func TestBootstrapRefreshUsesOrganizationDefinition(t *testing.T) {
	t.Parallel()
	for _, marker := range []bool{false, true} {
		theirs := model.Plan{ID: "org", Sections: []model.Section{{ID: "access", Bootstrap: marker}}}
		ours := model.Plan{Origin: model.Origin{PlanID: "org"}, Sections: []model.Section{{ID: "access", Bootstrap: !marker}}}
		got, _, problems := newRefresher().Refresh(model.Plan{}, theirs, ours, model.Origin{})
		if len(problems) != 0 || len(got.Sections) != 1 || got.Sections[0].Bootstrap != marker {
			t.Fatalf("organization marker not copied: %#v, %#v", got, problems)
		}
		if theirs.Sections[0].Bootstrap != marker || ours.Sections[0].Bootstrap == marker {
			t.Fatal("input marker changed")
		}
		theirs.ID = "different"
		refused, _, problems := newRefresher().Refresh(model.Plan{}, theirs, ours, model.Origin{})
		if len(problems) == 0 || len(refused.Sections) != 1 || refused.Sections[0].Bootstrap != !marker {
			t.Fatalf("refused refresh did not preserve local marker: %#v, %#v", refused, problems)
		}
	}
}
