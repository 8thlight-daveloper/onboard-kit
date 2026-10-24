package authoringvalidator_test

import "testing"

func TestSecondDifferentialGapRegressions(t *testing.T) {
	for _, item := range readCases[planCase](t, "testdata/second-differential-gaps.json") {
		t.Run(item.Name, func(t *testing.T) {
			assertProblems(t, newValidator().Validate(item.Plan), item.Want)
		})
	}
}
