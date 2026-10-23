package walker_test

import (
	"testing"

	contract "onboard-kit/contracts/section/walker"
	"onboard-kit/internal/model"
)

func TestHistoricalHashComparisonUsesExactEquality(t *testing.T) {
	t.Parallel()
	const digest = "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	for _, item := range []struct {
		name string
		hash string
		want contract.Comparison
	}{
		{"unknown", "", contract.ComparisonUnknown},
		{"exact", digest, contract.ComparisonSame},
		{"lexically-before-every-hex-digest", "!", contract.ComparisonDifferent},
		{"lexically-after-every-hex-digest", "~", contract.ComparisonDifferent},
		{"uppercase-is-not-repaired", "6C0301D995D0A3B0DAB95A641A087120C5266746E59C50D934C575E135CA2FDC", contract.ComparisonDifferent},
	} {
		t.Run(item.name, func(t *testing.T) {
			t.Parallel()
			record := model.Record{Outcomes: []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, SectionHash: item.hash}}}
			view, problems := newWalker().Walk(plan, record)
			if len(problems) != 0 || len(view.Outcomes) != 1 {
				t.Fatalf("missing history: %#v %#v", view, problems)
			}
			row := view.Outcomes[0]
			if row.CurrentHash != digest || row.Comparison != item.want || row.Outcome.SectionHash != item.hash || row.Outcome.Kind != model.OutcomeVerified {
				t.Fatalf("comparison changed held history or used lexical order: %#v", row)
			}
		})
	}
}
