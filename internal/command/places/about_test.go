package places

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"onboard-kit/internal/command/registry"
	"onboard-kit/internal/tui"
)

func TestTheStateHomeSaysHowItWasChosen(t *testing.T) {
	t.Setenv("OBK_HOME", "/from/setting")
	if home, err := StateHome(); err != nil || home != (Home{Path: "/from/setting", Chosen: ChosenByObkHome}) {
		t.Fatalf("got %+v, %v", home, err)
	}
	t.Setenv("OBK_HOME", "")
	t.Setenv("XDG_STATE_HOME", "/state")
	home, err := StateHome()
	if err != nil || filepath.Base(home.Path) != "obk" {
		t.Fatalf("got %+v, %v", home, err)
	}
	want := ChosenByDefault
	if runtime.GOOS != "darwin" {
		want = ChosenByXDGStateHome
	}
	if home.Chosen != want {
		t.Fatalf("chosen by %q, want %q", home.Chosen, want)
	}
}

func placeNamed(t *testing.T, places []tui.AboutPlace, name string) tui.AboutPlace {
	t.Helper()
	for _, place := range places {
		if place.Name == name {
			return place
		}
	}
	t.Fatalf("no place named %s", name)
	return tui.AboutPlace{}
}

func TestInTheEditorThePlanFolderAndRecordsRootAreThePlansOwn(t *testing.T) {
	listed := List(Context{Home: Home{Path: "/state/obk"}, PlanPath: "/work/acme/acme.obk.yaml", RecordsRoot: "/work/records"})
	folder := placeNamed(t, listed, "Plan folder")
	if folder.Path != "/work/acme" || !strings.HasPrefix(folder.Holds, "Holds acme.obk.yaml, ") {
		t.Fatalf("got %+v", folder)
	}
	if root := placeNamed(t, listed, "Records root"); root.Path != "/work/records" {
		t.Fatalf("got %+v", root)
	}
	if home := placeNamed(t, listed, "State home"); home.Chosen != "Chosen as the default for this system. Set OBK_HOME to choose another." {
		t.Fatalf("got %+v", home)
	}
}

func TestEveryPlaceSaysWhatItHoldsAndWhetherTheToolReadsOrWritesIt(t *testing.T) {
	for _, place := range List(Context{Home: Home{Path: "/state/obk", Chosen: ChosenByXDGStateHome}, WorkingFolder: "/work"}) {
		if place.Name == "" || !strings.HasPrefix(place.Holds, "Holds ") && place.Use != tui.PlaceNotUsedYet {
			t.Errorf("%+v", place)
		}
		if place.Path == "" {
			t.Errorf("%s has no path", place.Name)
		}
	}
}

func TestADamagedListIsAProblemAndAMissingOneIsNot(t *testing.T) {
	home := Home{Path: t.TempDir()}
	if problems := RegistryProblems(home); len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
	if err := os.WriteFile(registry.Path(home.Path), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	problems := RegistryProblems(home)
	if len(problems) != 1 || !strings.Contains(problems[0], "cannot be read: its first line is not onboard-kit plans 1.") {
		t.Fatalf("got %v", problems)
	}
	about := About("1.0.0", "1.0.0, set", Context{Home: home, HomeProblem: "No home."}, problems)
	if len(about.Problems) != 2 || about.Problems[0] != "No home." {
		t.Fatalf("got %v", about.Problems)
	}
}
