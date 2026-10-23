package refresher_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestRequiresRefreshUsesOrganizationDefinition(t *testing.T) {
	t.Parallel()
	for _, requires := range [][]string{nil, {}, {"b", "a", "b"}} {
		theirs := model.Plan{ID: "org", Sections: []model.Section{{ID: "access", Bootstrap: true, After: []string{"legacy"}, Requires: requires}}}
		ours := model.Plan{Origin: model.Origin{PlanID: "org"}, Sections: []model.Section{{ID: "access", Requires: []string{"local"}}}}
		base := model.Plan{Sections: []model.Section{{ID: "access", Requires: []string{"base"}}}}
		got, _, problems := newRefresher().Refresh(base, theirs, ours, model.Origin{})
		if len(problems) != 0 || !reflect.DeepEqual(got.Sections, theirs.Sections) {
			t.Fatalf("organization requires not copied: %#v, %#v", got.Sections, problems)
		}
		if !reflect.DeepEqual(theirs.Sections[0].Requires, requires) || !reflect.DeepEqual(ours.Sections[0].Requires, []string{"local"}) || !reflect.DeepEqual(base.Sections[0].Requires, []string{"base"}) {
			t.Fatal("input requires changed")
		}
		theirs.ID = "different"
		refused, _, problems := newRefresher().Refresh(base, theirs, ours, model.Origin{})
		if len(problems) != 1 || problems[0].Code != "refresh.different-plan" || !reflect.DeepEqual(refused, ours) {
			t.Fatalf("refused refresh changed local requires: %#v, %#v", refused, problems)
		}
	}
}
