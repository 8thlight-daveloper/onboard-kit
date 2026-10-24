package authoringvalidator_test

import "testing"

func TestDifferentialGapRegressions(t *testing.T) {
	for _, item := range readCases[planCase](t, "testdata/differential-gaps.json") {
		t.Run(item.Name, func(t *testing.T) {
			assertProblems(t, newValidator().Validate(item.Plan), item.Want)
		})
	}
}
