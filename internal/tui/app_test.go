package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
)

func acmePlan(t *testing.T) model.Plan {
	t.Helper()
	data, err := os.ReadFile("../../docs/reference/examples/acme/acme.obk.yaml")
	if err != nil {
		t.Fatal(err)
	}
	plan, problems := decoder.New().Decode(data)
	if len(problems) > 0 {
		t.Fatalf("the ACME plan has problems: %v", problems)
	}
	return plan
}

type recorder struct {
	saved []model.Plan
}

func (r *recorder) save(plan model.Plan) error {
	r.saved = append(r.saved, plan)
	return nil
}

func start(t *testing.T) (App, *recorder) {
	t.Helper()
	recorded := &recorder{}
	return NewApp(NewBackend(), acmePlan(t), recorded.save, model.AudienceOnboarder), recorded
}

func press(app App, keys ...tea.KeyMsg) App {
	for _, key := range keys {
		model, _ := app.Update(key)
		app = model.(App)
	}
	return app
}

func typed(text string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}
}

var (
	enter = tea.KeyMsg{Type: tea.KeyEnter}
	down  = tea.KeyMsg{Type: tea.KeyDown}
	space = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}
	esc   = tea.KeyMsg{Type: tea.KeyEsc}
)

func moveTo(t *testing.T, app App, id string) App {
	t.Helper()
	for index, question := range app.questions {
		if question.ID == id {
			app.cursor = index
			return app
		}
	}
	t.Fatalf("question %s is not asked", id)
	return app
}

func TestAnsweringATextQuestionAppliesAndSavesIt(t *testing.T) {
	t.Parallel()
	app, recorded := start(t)
	app = moveTo(t, app, "plan.organization")
	app = press(app, enter)
	if app.screen != screenEdit {
		t.Fatalf("expected the edit screen, got %v", app.screen)
	}
	app.edit.input.SetValue("")
	app = press(app, typed("Globex"), enter)
	if app.screen != screenQuestions {
		t.Fatalf("expected the question list again, got %v", app.screen)
	}
	if got := app.plan.Answers.Plan["organization"]; got.Text != "Globex" {
		t.Fatalf("answer not stored: %+v", got)
	}
	if len(recorded.saved) != 1 {
		t.Fatalf("expected one save, got %d", len(recorded.saved))
	}
}

func TestAChoiceIsPickedFromItsOptions(t *testing.T) {
	t.Parallel()
	app, recorded := start(t)
	app = moveTo(t, app, "plan.discipline")
	app = press(app, enter)
	if len(app.edit.choices) == 0 {
		t.Fatal("a choice question should offer a list")
	}
	labels := make([]string, 0, len(app.edit.choices))
	for _, option := range app.edit.choices {
		labels = append(labels, option.label)
	}
	if strings.Join(labels, "|") != "software|design|deferred, answer later|clear the answer" {
		t.Fatalf("unexpected choices %q", labels)
	}
	app = press(app, down, enter)
	if got := app.plan.Answers.Plan["discipline"]; got.Text != "design" {
		t.Fatalf("expected design, got %+v", got)
	}
	if len(recorded.saved) != 1 {
		t.Fatalf("expected one save, got %d", len(recorded.saved))
	}
}

func TestDeferringAndClearingChangeTheAnswer(t *testing.T) {
	t.Parallel()
	app, _ := start(t)
	app = moveTo(t, app, "access.access_request_url")
	app = press(app, typed("d"))
	if got := app.plan.Answers.Sections["access"]["access_request_url"]; got.Kind != model.ValueKindDeferred {
		t.Fatalf("expected a deferred answer, got %+v", got)
	}
	app = press(app, typed("x"))
	if _, there := app.plan.Answers.Sections["access"]["access_request_url"]; there {
		t.Fatal("the answer should be cleared")
	}
}

func TestASecretIsRefusedAndTheEditStays(t *testing.T) {
	t.Parallel()
	app, recorded := start(t)
	app = moveTo(t, app, "access.repo_url")
	before := app.plan.Answers.Sections["access"]["repo_url"]
	app = press(app, enter)
	app.edit.input.SetValue("")
	app = press(app, typed("https://user:hunter2@git.example.test/r.git"), enter)
	if app.screen != screenEdit || len(app.edit.problems) == 0 {
		t.Fatalf("expected to stay in the edit with a problem, got screen %v problems %v", app.screen, app.edit.problems)
	}
	if got := app.plan.Answers.Sections["access"]["repo_url"]; got.Text != before.Text {
		t.Fatalf("the answer changed to %+v", got)
	}
	if len(recorded.saved) != 0 {
		t.Fatal("a refused answer must not be saved")
	}
	app = press(app, esc)
	if app.screen != screenQuestions {
		t.Fatal("escape should return to the list")
	}
}

func TestSectionsAreSwitchedOffAndOn(t *testing.T) {
	t.Parallel()
	app, recorded := start(t)
	app = press(app, typed("s"))
	if app.screen != screenSections {
		t.Fatalf("expected the sections screen, got %v", app.screen)
	}
	app = press(app, space)
	if isOn(app.plan.Sections[0]) {
		t.Fatal("the first section should be off")
	}
	app = press(app, space)
	if !isOn(app.plan.Sections[0]) {
		t.Fatal("the first section should be on again")
	}
	if len(recorded.saved) != 2 {
		t.Fatalf("expected two saves, got %d", len(recorded.saved))
	}
}

func TestEveryScreenDrawsSomething(t *testing.T) {
	t.Parallel()
	app, _ := start(t)
	screens := map[string]App{
		"questions": app,
		"edit":      press(moveTo(t, app, "plan.organization"), enter),
		"sections":  press(app, typed("s")),
		"review":    press(app, typed("r")),
	}
	for name, shown := range screens {
		if strings.TrimSpace(shown.View()) == "" {
			t.Errorf("the %s screen is empty", name)
		}
	}
	if !strings.Contains(screens["review"].View(), "ACME Corp") {
		t.Error("the review should show the assembled document")
	}
}

func TestMovingStaysInsideTheList(t *testing.T) {
	t.Parallel()
	app, _ := start(t)
	app = press(app, typed("k"), typed("k"))
	if app.cursor != 0 {
		t.Fatalf("cursor went to %d", app.cursor)
	}
	app = press(app, typed("G"), down, down)
	if app.cursor != len(app.questions)-1 {
		t.Fatalf("cursor went to %d of %d", app.cursor, len(app.questions))
	}
}

func TestQuittingFinishesTheProgram(t *testing.T) {
	t.Parallel()
	app, _ := start(t)
	app = press(app, typed("q"))
	if !app.Finished() {
		t.Fatal("q should finish")
	}
}

func TestWordsAreSplitAndJoinedWithQuotes(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"make test":             {"make", "test"},
		`git commit -m "a b c"`: {"git", "commit", "-m", "a b c"},
		`run '' two`:            {"run", "", "two"},
		"  spaced   out  ":      {"spaced", "out"},
	}
	for text, want := range cases {
		got := splitWords(text)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("splitWords(%q) = %q, want %q", text, got, want)
		}
		again := splitWords(joinWords(got))
		if strings.Join(again, "|") != strings.Join(want, "|") {
			t.Errorf("joinWords round trip of %q gave %q", text, again)
		}
	}
}

func TestPeopleAreSeparatedBySemicolons(t *testing.T) {
	t.Parallel()
	question := model.Question{Type: model.VariableTypePerson, Multiple: true}
	value := parseEntry(question, "Ann Lee, ann@x.test, @ann; Bo Ray, bo@x.test, @bo ;")
	if len(value.List) != 2 || value.List[1] != "Bo Ray, bo@x.test, @bo" {
		t.Fatalf("unexpected people %q", value.List)
	}
	if parseEntry(question, "   ").Kind != model.ValueKindNone {
		t.Fatal("blank text means no answer")
	}
}
