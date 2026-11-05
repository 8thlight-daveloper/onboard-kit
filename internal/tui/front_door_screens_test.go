package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func frontDoorScreens() map[string]func(FrontDoor) FrontDoor {
	keys := func(pressed ...tea.Msg) func(FrontDoor) FrontDoor {
		return func(door FrontDoor) FrontDoor { return pressDoor(door, pressed...) }
	}
	return map[string]func(FrontDoor) FrontDoor{
		"front-door-first":       keys(),
		"front-door-first-begin": keys(down),
		"front-door-help":        keys(typed("?")),
		"front-door-new":         keys(enter),
		"front-door-new-confirm": keys(enter, enter),
		"front-door-begin":       keys(down, enter),
		"front-door-about":       keys(typed("a")),
	}
}

func frontDoorPlanScreens() map[string]func(FrontDoor) FrontDoor {
	keys := func(pressed ...tea.Msg) func(FrontDoor) FrontDoor {
		return func(door FrontDoor) FrontDoor { return pressDoor(door, pressed...) }
	}
	return map[string]func(FrontDoor) FrontDoor{
		"front-door-plans":        keys(),
		"front-door-plans-action": keys(typed("G"), up),
		"front-door-missing":      keys(down, down),
		"front-door-forget":       keys(down, down, typed("x")),
		"front-door-open":         keys(typed("G"), up, enter, typed("notes.txt"), enter),
	}
}

func TestFrontDoorScreensMatchTheirWireframes(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(profileAscii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
	builds := map[string]func(goldenSize) FrontDoor{}
	for name, steps := range frontDoorScreens() {
		builds[name] = func(size goldenSize) FrontDoor {
			return steps(pressDoor(NewFrontDoor(newFakeFrontDoorPort()), tea.WindowSizeMsg{Width: size.width, Height: size.height}))
		}
	}
	for name, steps := range frontDoorPlanScreens() {
		builds[name] = func(size goldenSize) FrontDoor {
			return steps(pressDoor(NewFrontDoor(newFakeFrontDoorPort(knownPlans()...)), tea.WindowSizeMsg{Width: size.width, Height: size.height}))
		}
	}
	builds["front-door-damaged"] = func(size goldenSize) FrontDoor {
		port := newFakeFrontDoorPort()
		port.listProblem = "The list of plans at /home/sam/.local/state/obk/plans.txt cannot be read: line 2 is not a time, a tab, and a full path. It is left as it is, and plans you open are not added to it. Remove the file to start a new list."
		return pressDoor(NewFrontDoor(port), tea.WindowSizeMsg{Width: size.width, Height: size.height})
	}
	for name, build := range builds {
		for _, size := range goldenSizes {
			compareWireframe(t, filepath.Join("testdata", "screens", fmt.Sprintf("%s.%d.txt", name, size.width)), doorWireframe(build(size)))
		}
	}
}

func doorWireframe(door FrontDoor) string {
	return wireframeOf(doorView(door))
}

func compareWireframe(t *testing.T, path string, got string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s is missing, run go test ./internal/tui -run Wireframes -update: %v", path, err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the screen.\nwant:\n%s\ngot:\n%s", path, want, got)
	}
}
