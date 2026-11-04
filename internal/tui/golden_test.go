package tui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

var update = flag.Bool("update", false, "rewrite the screen files under testdata/screens")

type goldenSize struct {
	width  int
	height int
}

var goldenSizes = []goldenSize{{80, 24}, {120, 30}, {140, 40}}

func goldenScreens(t *testing.T) map[string]func(size goldenSize) App {
	t.Helper()
	build := func(plan model.Plan, audience model.Audience, steps ...func(App) App) func(goldenSize) App {
		return func(size goldenSize) App {
			app, _ := startWith(t, plan, audience)
			app = resize(app, size.width, size.height)
			for _, step := range steps {
				app = step(app)
			}
			return app
		}
	}
	keys := func(pressed ...tea.KeyMsg) func(App) App {
		return func(app App) App { return press(app, pressed...) }
	}
	goTo := func(id string) func(App) App {
		return func(app App) App { return moveTo(t, app, id) }
	}
	return map[string]func(goldenSize) App{
		"walk":                         build(blankPlan(t), model.AudienceOnboarder, goTo("plan.team_state")),
		"edit":                         build(blankPlan(t), model.AudienceOnboarder, keys(typed("Glob"))),
		"overview":                     build(acmePlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("o"))),
		"review":                       build(acmePlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("d"), typed("r"))),
		"review-left-over":             build(leftOverPlan(t), model.AudienceOnboarder),
		"review-prune":                 build(leftOverPlan(t), model.AudienceOnboarder, keys(typed("h"))),
		"sections":                     build(blankPlan(t), model.AudienceOnboarder, goTo("plan.team_state"), keys(typed("s"))),
		"onboardee":                    build(blankPlan(t), model.AudienceOnboardee),
		"example":                      buildFake(),
		"pick":                         build(blankPlan(t), model.AudienceOnboarder, goTo("access.onboarding_owner_role")),
		"pick-many":                    build(blankPlan(t), model.AudienceOnboarder, goTo("stuck-general.general_contacts"), keys(typed("j"), space, typed("k"), space)),
		"pick-empty":                   build(withoutDirectory(blankPlan(t)), model.AudienceOnboarder, goTo("access.onboarding_owner_role")),
		"pick-empty-onboardee":         build(withoutDirectory(blankPlan(t)), model.AudienceOnboardee),
		"subsections":                  buildSubsections(),
		"contacts-list":                build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"))),
		"contacts-roles":               build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"), tab)),
		"contacts-teams":               build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"), tab, tab)),
		"contacts-form":                build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"), typed("a"), typed("Dana Lee"), enter, typed("dana@acme.example"))),
		"contacts-holders":             build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"), tab, typed("a"), typed("Release lead"), enter, down, space, down, space)),
		"contacts-override":            build(contactsWithFlags(t), model.AudienceOnboarder, keys(typed("c"), tab, tab, typed("a"), typed("Data team"), enter, down, enter, down, space)),
		"contacts-team-form":           build(acmePlan(t), model.AudienceOnboarder, keys(typed("c"), tab, tab, enter)),
		"contacts-team-contact":        build(acmePlan(t), model.AudienceOnboarder, keys(typed("c"), tab, tab, enter, enter, enter)),
		"contacts-remove-team-contact": build(acmePlan(t), model.AudienceOnboarder, keys(typed("c"), down, down, down, typed("x"))),
		"contacts-remove":              build(acmePlan(t), model.AudienceOnboarder, keys(typed("c"), down, typed("x"))),
		"contacts-remove-role":         build(acmePlan(t), model.AudienceOnboarder, keys(typed("c"), tab, typed("x"))),
		"contacts-empty":               build(withoutDirectory(blankPlan(t)), model.AudienceOnboarder, keys(esc, typed("c"))),
		"preparation-people":           preparing(t, "preparation-people"),
		"preparation-add":              preparing(t, "preparation-add"),
		"preparation-person":           preparing(t, "preparation-person"),
		"preparation-recipient":        preparing(t, "preparation-recipient"),
		"preparation-pick":             preparing(t, "preparation-pick"),
		"preparation-arrangement":      preparing(t, "preparation-arrangement"),
		"preparation-helpers":          preparing(t, "preparation-helpers"),
		"preparation-problem":          preparing(t, "preparation-problem"),
		"preparation-remove":           preparing(t, "preparation-remove"),
		"preparation-drop":             preparing(t, "preparation-drop"),
		"preparation-review":           preparing(t, "preparation-review"),
		"preparation-guide":            preparing(t, "preparation-guide"),
		"invitation-people":            inviting(t, "invitation-people"),
		"invitation-blocked":           inviting(t, "invitation-blocked"),
		"invitation-form":              inviting(t, "invitation-form"),
		"invitation-pick":              inviting(t, "invitation-pick"),
		"invitation-helpers":           inviting(t, "invitation-helpers"),
		"invitation-missing":           inviting(t, "invitation-missing"),
		"invitation-confirm":           inviting(t, "invitation-confirm"),
		"invitation-written":           inviting(t, "invitation-written"),
		"too-small": func(size goldenSize) App {
			app, _ := startWith(t, blankPlan(t), model.AudienceOnboarder)
			return resize(app, size.width, 10)
		},
	}
}

func leftOverPlan(t *testing.T) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	plan.Answers.Plan["old_tool"] = model.Value{Kind: model.ValueKindText, Text: "Trello"}
	plan.Answers.Plan["old_token"] = model.Value{Kind: model.ValueKindText, Text: "sk-live-4242424242424242424242"}
	plan.Answers.Sections = map[string]map[string]model.Value{"retired": {"flag": {Kind: model.ValueKindFlag, Flag: true}}}
	return plan
}

func buildFake() func(goldenSize) App {
	return func(size goldenSize) App {
		backend := newFakeBackend(
			textQuestion("plan.organization", "plan", "Name of the organization", "Acme Industrial"),
			textQuestion("plan.lead", "plan", "Who leads the onboarding, and what do they need to tell a newcomer on the first day?", "Sam Rivera, sam@acme.test"),
		)
		return resize(startFake(backend), size.width, size.height)
	}
}

func TestScreensMatchTheirWireframes(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	for name, build := range goldenScreens(t) {
		for _, size := range goldenSizes {
			path := filepath.Join("testdata", "screens", fmt.Sprintf("%s.%d.txt", name, size.width))
			got := wireframe(build(size))
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%s is missing, run go test ./internal/tui -run Wireframes -update: %v", path, err)
			}
			if string(want) != got {
				t.Errorf("%s differs from the screen.\nwant:\n%s\ngot:\n%s", path, want, got)
			}
		}
	}
}

func wireframe(app App) string {
	rows := viewLines(app)
	for index, row := range rows {
		rows[index] = strings.TrimRight(row, " ")
	}
	return strings.Join(rows, "\n") + "\n"
}

func contactsWithFlags(t *testing.T) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	plan.Directory.Contacts = append(plan.Directory.Contacts, model.Contact{ID: "pat-lee", Name: "Pat Lee", Email: "pat.lee.acme.example"}, model.Contact{ID: "front-desk", Name: "Front desk"})
	plan.Directory.Roles = append(plan.Directory.Roles, model.Role{ID: "security-lead", Name: "Security lead"})
	return plan
}

func withoutDirectory(plan model.Plan) model.Plan {
	plan.Directory = model.Directory{}
	return plan
}

func subsectionBackend() (*fakeBackend, model.Plan) {
	backend := newFakeBackend(
		textQuestion("plan.organization", "plan", "Name of the organization", ""),
		textQuestion("access.owner", "access", "Who owns access?", ""),
		textQuestion("access-request.url", "access-request", "Where do people request access?", ""),
		textQuestion("access-tokens.rotation", "access-tokens", "How often are tokens rotated?", ""),
		textQuestion("tools.editor", "tools", "Which editor is standard?", ""),
	)
	plan := sectionsWithParents("access", "", "access-request", "access", "access-tokens", "access-request", "tools", "")
	return backend, plan
}

func buildSubsections() func(goldenSize) App {
	return func(size goldenSize) App {
		backend, plan := subsectionBackend()
		backend.answers["access-request.url"] = model.Value{Kind: model.ValueKindText, Text: "https://access.example.test"}
		app := resize(startFakeIn(backend, plan, model.AudienceOnboarder), size.width, size.height)
		return app.openQuestion(2).openOverview()
	}
}
