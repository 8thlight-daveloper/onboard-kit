package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func headerOf(app App) string { return strings.TrimSpace(viewLines(app)[0]) }

func TestWithNoColourTheCurrentTabIsInBracketsAndTheOthersCarryTheirKeys(t *testing.T) {
	withoutColour(t)
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = resize(app, 120, 30)
	cases := map[string]App{
		"Onboarding plan  Walk w  Overview o  [Review]  Sections s": app,
		"Onboarding plan  [Walk]  Overview o  Review r  Sections s": moveTo(t, app, "plan.team_state"),
		"Onboarding plan  Walk w  [Overview]  Review r  Sections s": app.openOverview(),
		"Onboarding plan  Walk w  Overview o  Review r  [Sections]": press(app, typed("s")),
	}
	for want, shown := range cases {
		if got := headerOf(shown); !strings.HasPrefix(got, want) {
			t.Errorf("header is %q, want it to start with %q", got, want)
		}
	}
}

func TestWithColourTheCurrentTabIsBoldAccentAndTheOthersAreFaint(t *testing.T) {
	withColour(t)
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	view := resize(app.openOverview(), 120, 30).View()
	if strings.Contains(plainText(view), "[Overview]") {
		t.Error("colour replaces the brackets")
	}
	expectBright(t, view, "Overview")
	if got := sgrBefore(t, view, "Overview"); !strings.Contains(got, "1;") || !strings.Contains(got, "38;5;32") {
		t.Errorf("the current tab is bold in the accent, got SGR %q", got)
	}
	expectFaint(t, view, "Walk w")
	expectFaint(t, view, "Review r")
	expectFaint(t, view, "Sections s")
}

func TestTheOnboardeeHasNoSectionsTab(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	for _, shown := range []App{app, app.openOverview(), app.openReview()} {
		header := headerOf(resize(shown, 120, 30))
		if strings.Contains(header, "Sections") || !strings.HasPrefix(header, "Your onboarding") {
			t.Errorf("the onboardee header is %q", header)
		}
	}
}

func TestWhileTypingTheTabsShowNoKeysBecauseLettersAreText(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	header := headerOf(resize(app, 120, 30))
	if !strings.HasPrefix(header, "Onboarding plan  [Walk]  Overview  Review  Sections") {
		t.Fatalf("header is %q", header)
	}
	choosing := moveTo(t, app, "plan.team_state")
	if !strings.Contains(headerOf(resize(choosing, 120, 30)), "Overview o") {
		t.Error("keys show once the person is choosing")
	}
}

func TestTheHeaderDropsTheProgressThenTheKeysBeforeItOverflows(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	for _, width := range widths {
		header := headerOf(resize(app, width, 30))
		for _, name := range []string{"Walk", "Overview", "Review", "Sections"} {
			if !strings.Contains(header, name) {
				t.Errorf("%d columns: %q lacks %s", width, header, name)
			}
		}
	}
	if !strings.Contains(headerOf(resize(app, 80, 30)), "answered") {
		t.Error("80 columns fits the progress")
	}
	narrow := headerOf(resize(app, 60, 30))
	if strings.Contains(narrow, "answered") || strings.Contains(narrow, "Review r") {
		t.Errorf("60 columns keeps only the names: %q", narrow)
	}
}

func TestWhenTheHeaderHasNoKeysTheHintLineCarriesThem(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	choosing := moveTo(t, app, "plan.team_state")
	wide := viewLines(resize(choosing, 120, 30))
	if last := wide[len(wide)-1]; strings.Contains(last, "o overview") {
		t.Errorf("the header carries the keys at 120 columns: %q", last)
	}
	narrow := viewLines(resize(choosing, 60, 30))
	if last := narrow[len(narrow)-1]; !strings.Contains(last, "o overview") {
		t.Errorf("the hint line carries the keys at 60 columns: %q", last)
	}
}

func TestTheScreenKeysWorkFromTheSectionsScreen(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	sections := press(moveTo(t, app, "plan.team_state"), typed("s"))
	for key, want := range map[string]screen{"w": screenWalk, "o": screenOverview, "r": screenReview} {
		if got := press(sections, typed(key)).screen; got != want {
			t.Errorf("%s from the sections screen opened %v, want %v", key, got, want)
		}
	}
	if press(sections, typed("w")).questions[press(sections, typed("w")).cursor].ID != "plan.team_state" {
		t.Error("w returns to the question the person left")
	}
}

func TestTheTabsNameTheScreenWhileHelpIsOpen(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	header := headerOf(resize(press(app.openOverview(), typed("?")), 120, 30))
	if !strings.Contains(header, "[Overview]") {
		t.Errorf("header is %q", header)
	}
}
