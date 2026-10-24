//go:build darwin || linux

package editor_test

import (
	"strings"
	"testing"
)

func TestRawAdmissionAndActualLocatorBindBeforeAction(t *testing.T) {
	for _, item := range []struct{ name, plan, code string }{
		{"raw-body-before-decoder", strings.Repeat("x", 1048577), "edit.limit"},
		{"actual-locator-disagrees", "resultsPath: other/{account}/custom.yaml\n" + planBody, "edit.context"},
	} {
		t.Run(item.name, func(t *testing.T) {
			s := session(t, item.plan, recordBody)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, confirm+"UNREAD_ACTION\n")
			report(t, got, item.code, 1, false)
			if got.consumed != len(confirm) || clock.calls != 0 {
				t.Fatal("admission read ahead or sampled runtime event")
			}
			unchanged(t, s, before)
		})
	}
}
