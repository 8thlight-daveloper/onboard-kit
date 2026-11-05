package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

func TestTheExampleIsDimPlaceholderTextInTheInput(t *testing.T) {
	withColour(t)
	backend := newFakeBackend(textQuestion("plan.organization", "plan", "Name of the organization", "Acme Industrial"))
	app := resize(startFake(backend), 120, 30)
	if !strings.Contains(app.entry.input.Placeholder, "Acme Industrial") {
		t.Fatalf("placeholder is %q", app.entry.input.Placeholder)
	}
	view := app.View()
	expectFaint(t, view, "Acme Industrial")
	typedApp := press(app, typed("Globex"))
	if strings.Contains(plainView(typedApp), "e.g. Acme") {
		t.Error("the placeholder goes away once the person types")
	}
}

func TestTheExampleIsALineInTheNotesAtEveryWidth(t *testing.T) {
	backend := newFakeBackend(textQuestion("plan.organization", "plan", "Name of the organization", "Acme Industrial"))
	app := startFake(backend)
	for _, width := range []int{80, 120, 140} {
		if !strings.Contains(plainView(resize(app, width, 30)), "Example: Acme Industrial") {
			t.Errorf("%d columns lacks the example line", width)
		}
	}
}

func TestTheExampleIsNeverAnAnswer(t *testing.T) {
	backend := newFakeBackend(
		textQuestion("plan.organization", "plan", "Name of the organization", "Acme Industrial"),
		textQuestion("plan.lead", "plan", "Who leads", ""),
	)
	app := resize(startFake(backend), 120, 30)
	app = press(app, enter)
	if len(backend.submitted) != 0 {
		t.Fatalf("enter on an empty box must not save the example, saved %v", backend.submitted)
	}
	if app.cursor != 1 {
		t.Fatalf("enter on an empty unanswered box moves on, cursor %d", app.cursor)
	}
	if strings.Contains(plainView(app), "Example:") {
		t.Error("a question with no example shows no example line")
	}
}

func TestALongExampleIsCutToTheInputWidth(t *testing.T) {
	long := strings.Repeat("example ", 40)
	backend := newFakeBackend(textQuestion("plan.organization", "plan", "Name", long))
	for _, width := range widths {
		for _, row := range viewLines(resize(startFake(backend), width, 30)) {
			if lipgloss.Width(row) > width {
				t.Fatalf("%d columns: row is %d wide", width, lipgloss.Width(row))
			}
		}
	}
}

func TestGuidedAnsweringOfManyQuestionsShowsThePositionAndSection(t *testing.T) {
	var questions []model.Question
	for index := 0; index < 11; index++ {
		questions = append(questions, textQuestion("plan.q"+strings.Repeat("a", index), "plan", "Question text", ""))
	}
	backend := newFakeBackend(questions...)
	app := resize(startFake(backend), 120, 30)
	app = press(app, typed("one"), enter, typed("two"), enter, typed("three"), enter)
	view := plainView(app)
	for _, want := range []string{"4 of 11", "Plan", "3 of 11 answered"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "✓✓✓◆") {
		t.Errorf("the progress rail should show three answered then the position:\n%s", view)
	}
}
