package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func TestANoticeShowsInTheStatusLineUntilTheFirstKey(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	notice := "Refreshed from the organization's plan: 2 new questions"
	app = resize(app.WithNotice(notice), 120, 30)
	if !strings.Contains(app.View(), notice) {
		t.Fatalf("the notice is not shown:\n%s", app.View())
	}
	app = press(app, typed("?"))
	app = press(app, typed("x"))
	if strings.Contains(app.View(), notice) {
		t.Fatalf("the notice stays after a key:\n%s", app.View())
	}
}

func TestAnEmptyNoticeChangesNothing(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	if app.WithNotice("").status != (line{}) {
		t.Fatal("an empty notice set the status line")
	}
}
