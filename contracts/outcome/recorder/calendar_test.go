package recorder_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
)

func TestEveryMonthLastDayAcceptedAsInjected(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		on   model.Timestamp
	}{
		{"january", "2026-01-31T12:34:56Z"},
		{"february-common", "2026-02-28T12:34:56Z"},
		{"march", "2026-03-31T12:34:56Z"},
		{"april", "2026-04-30T12:34:56Z"},
		{"may", "2026-05-31T12:34:56Z"},
		{"june", "2026-06-30T12:34:56Z"},
		{"july", "2026-07-31T12:34:56Z"},
		{"august", "2026-08-31T12:34:56Z"},
		{"september", "2026-09-30T12:34:56Z"},
		{"october", "2026-10-31T12:34:56Z"},
		{"november", "2026-11-30T12:34:56Z"},
		{"december", "2026-12-31T12:34:56Z"},
		{"leap-century-divisible-by-400", "2400-02-29T12:34:56Z"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
			want := model.Record{Outcomes: []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, On: c.on, SectionHash: "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"}}}
			got, problems := newRecorder().Record(plan, model.Record{}, contract.Proposal{Section: "access", Kind: model.OutcomeVerified}, c.on)
			if len(problems) != 0 || !reflect.DeepEqual(got, want) {
				t.Fatalf("valid last day %q changed or refused: %#v %#v", c.on, got, problems)
			}
		})
	}
}

func TestEveryMonthNextNumericDayRefusedWithoutRepair(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		on   model.Timestamp
	}{
		{"january", "2026-01-32T12:34:56Z"},
		{"february-common", "2026-02-29T12:34:56Z"},
		{"march", "2026-03-32T12:34:56Z"},
		{"april", "2026-04-31T12:34:56Z"},
		{"may", "2026-05-32T12:34:56Z"},
		{"june", "2026-06-31T12:34:56Z"},
		{"july", "2026-07-32T12:34:56Z"},
		{"august", "2026-08-32T12:34:56Z"},
		{"september", "2026-09-31T12:34:56Z"},
		{"october", "2026-10-32T12:34:56Z"},
		{"november", "2026-11-31T12:34:56Z"},
		{"december", "2026-12-32T12:34:56Z"},
		{"february-leap", "2024-02-30T12:34:56Z"},
		{"february-leap-century", "2400-02-30T12:34:56Z"},
		{"february-nonleap-century", "2100-02-29T12:34:56Z"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
			want := []model.Problem{{Code: "outcome.timestamp-invalid", Severity: model.SeverityError, Message: "Supply a valid UTC outcome timestamp.", Detail: map[string]string{"field": "on"}}}
			got, problems := newRecorder().Record(plan, model.Record{}, contract.Proposal{Section: "access", Kind: model.OutcomeVerified}, c.on)
			if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(problems, want) {
				t.Fatalf("invalid next day %q accepted or repaired: %#v %#v", c.on, got, problems)
			}
		})
	}
}
