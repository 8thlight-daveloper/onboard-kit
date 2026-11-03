package runner_test

import (
	"strings"
	"testing"
	"time"

	"onboard-kit/acceptance/runner"
)

func shellScenario(script string, terminal string, expect string) string {
	return "name: terminal\nsteps:\n  - run: [/bin/sh, -c, '" + script + "']\n    terminal:\n" + terminal + "\n    expect:\n" + expect + "\n"
}

const enterScreen = `printf "\033[?1049h\033[?25l\033[?2004h"`
const leaveScreen = `printf "\033[?1049l\033[?25h\033[?2004l"`

func failureText(failures []runner.Failure) string {
	var text []string
	for _, failure := range failures {
		text = append(text, failure.String())
	}
	return strings.Join(text, "\n")
}

func TestTerminalStepWaitsPressesAndChecksModesAndExit(t *testing.T) {
	t.Parallel()
	script := enterScreen + `; printf hello; read answer; ` + leaveScreen + `; exit 3`
	terminal := `      size: {columns: 60, rows: 10}
      actions:
        - wait-for: hello
        - expect-modes: {alternate_screen: true, cursor_visible: false, bracketed_paste: true}
        - press: Enter
        - wait-exit: true
        - expect-modes: {alternate_screen: false, cursor_visible: true, bracketed_paste: false}`
	if failures := runScenario(t, shellScenario(script, terminal, "      exit: 3")); len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
}

func TestTerminalStepReportsAWrongMode(t *testing.T) {
	t.Parallel()
	script := enterScreen + `; printf hello; sleep 3`
	terminal := "      actions:\n        - wait-for: hello\n        - expect-modes: {alternate_screen: false, cursor_visible: false}"
	text := failureText(runScenario(t, shellScenario(script, terminal, "")))
	if !strings.Contains(text, "alternate_screen expected false, got true") || strings.Contains(text, "cursor_visible") {
		t.Fatalf("only the wrong mode should be named, got %q", text)
	}
}

func TestTerminalStepReportsAWrongExitCode(t *testing.T) {
	t.Parallel()
	terminal := "      actions:\n        - wait-exit: true"
	failures := runScenario(t, shellScenario("exit 3", terminal, "      exit: 0"))
	if !strings.Contains(failureText(failures), "expected 0, got 3") {
		t.Fatalf("the wrong exit code should be named, got %q", failureText(failures))
	}
}

func TestTerminalStepFailureShowsTheScreen(t *testing.T) {
	t.Parallel()
	terminal := "      actions:\n        - wait-for: beta"
	options := runner.Options{TimeLimit: 2 * time.Second}
	scenario, err := runner.Parse([]byte(shellScenario("printf alpha; sleep 5", terminal, "      exit: 0")))
	if err != nil {
		t.Fatal(err)
	}
	options.BinaryPath = fakeBinaryPath
	options.RepositoryRoot = t.TempDir()
	text := failureText(runner.Run(scenario, options))
	for _, want := range []string{`terminal action 1, wait-for "beta"`, "the screen at the failure", "alpha"} {
		if !strings.Contains(text, want) {
			t.Errorf("failure text should hold %q, got %q", want, text)
		}
	}
}

func TestTerminalStepSendsRealSignals(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		script string
		signal string
		exit   string
	}{
		"TERM is handled": {`trap "exit 4" TERM; printf ready; while :; do sleep 0.1; done`, "TERM", "4"},
		"HUP ends it":     {`printf ready; read answer`, "HUP", "129"},
		"INT ends it":     {`printf ready; read answer`, "INT", "130"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			terminal := "      actions:\n        - wait-for: ready\n        - signal: " + test.signal + "\n        - wait-exit: true"
			if failures := runScenario(t, shellScenario(test.script, terminal, "      exit: "+test.exit)); len(failures) > 0 {
				t.Fatal(failureText(failures))
			}
		})
	}
}

func TestTerminalColorsAcceptTheDefaultAndTheSixteen(t *testing.T) {
	t.Parallel()
	script := `printf "\033[32mgreen\033[0m \033[1;31mred\033[0m \033[92mbright\033[0m \033[2mfaint\033[0m"`
	terminal := "      actions:\n        - wait-for: faint\n        - expect-colors: terminal-only"
	if failures := runScenario(t, shellScenario(script+"; sleep 3", terminal, "")); len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
}

func TestTerminalColorsRefuseAFixedColor(t *testing.T) {
	t.Parallel()
	for name, escape := range map[string]string{"256 color": `\033[38;5;196m`, "true color": `\033[38;2;1;2;3m`, "background": `\033[48;5;100m`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			script := `printf "` + escape + `fixed\033[0m"; sleep 3`
			terminal := "      actions:\n        - wait-for: fixed\n        - expect-colors: terminal-only"
			failures := runScenario(t, shellScenario(script, terminal, ""))
			text := failureText(failures)
			if !strings.Contains(text, "expect-colors") || !strings.Contains(text, "neither the default color") {
				t.Fatalf("a fixed color should fail, got %q", text)
			}
		})
	}
}

func TestTerminalPlainAcceptsUnstyledTextAndRefusesStyleOrColor(t *testing.T) {
	t.Parallel()
	plain := "      actions:\n        - wait-for: word\n        - expect-colors: plain"
	if failures := runScenario(t, shellScenario(`printf "plain word"; sleep 3`, plain, "")); len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
	for name, escape := range map[string]string{"bold": `\033[1m`, "faint": `\033[2m`, "reverse": `\033[7m`, "underline": `\033[4m`, "ANSI color": `\033[32m`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			text := failureText(runScenario(t, shellScenario(`printf "`+escape+`word\033[0m"; sleep 3`, plain, "")))
			if !strings.Contains(text, "should be plain") {
				t.Fatalf("%s should fail the plain check, got %q", name, text)
			}
		})
	}
}

func TestTerminalStepTypesResizesAndPassesTheEnvironment(t *testing.T) {
	t.Parallel()
	script := `printf "TERM=$TERM FOO=$FOO\n"; read line; printf "got:$line\n"; trap "stty size" WINCH; while :; do sleep 0.1; done`
	terminal := `      env: {FOO: bar}
      size: {columns: 50, rows: 12}
      actions:
        - wait-for: "TERM=xterm-256color FOO=bar"
        - type: abc
        - press: Enter
        - wait-for: "got:abc"
        - resize: {columns: 70, rows: 20}
        - wait-for: "20 70"`
	if failures := runScenario(t, shellScenario(script, terminal, "")); len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
}

func TestTerminalStepStillChecksFiles(t *testing.T) {
	t.Parallel()
	terminal := "      actions:\n        - wait-exit: true"
	expect := "      exit: 0\n      files:\n        out.txt: {same-as: want.txt}"
	failures := runScenarioWith(t, shellScenario("echo hi > out.txt", terminal, expect), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"want.txt": "hi\n"})
	if len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
}

func TestTerminalStepSchemaErrors(t *testing.T) {
	t.Parallel()
	base := "name: terminal\nsteps:\n  - run: [/bin/true]\n"
	cases := map[string]string{
		"unknown action field":    base + "    terminal:\n      actions:\n        - sleep: 1\n",
		"two fields in an action": base + "    terminal:\n      actions:\n        - {press: q, type: x}\n",
		"no field in an action":   base + "    terminal:\n      actions:\n        - {}\n",
		"talk":                    base + "    talk: []\n    terminal: {}\n",
		"stdin":                   base + "    stdin: x\n    terminal: {}\n",
		"stdout expectation":      base + "    terminal: {}\n    expect:\n      stdout: {contains: x}\n",
		"zero columns":            base + "    terminal:\n      size: {columns: 0, rows: 5}\n",
		"bad signal":              base + "    terminal:\n      actions:\n        - signal: KILL\n",
		"bad mode":                base + "    terminal:\n      actions:\n        - expect-modes: {mouse: true}\n",
		"bad colors value":        base + "    terminal:\n      actions:\n        - expect-colors: any\n",
		"bad resize":              base + "    terminal:\n      actions:\n        - resize: {columns: 0, rows: 5}\n",
		"unknown terminal field":  base + "    terminal:\n      shell: bash\n",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := runner.Parse([]byte(text)); err == nil {
				t.Fatal("the scenario should not parse")
			}
		})
	}
}

func TestTerminalStepBlanksHostVariablesThatChangeHowTheProgramDraws(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("NO_COLOR", "1")
	t.Setenv("EDITOR", "vi")
	t.Setenv("LANG", "en_US.UTF-8")
	script := `printf "COLORTERM=[$COLORTERM] NO_COLOR=[$NO_COLOR] EDITOR=[$EDITOR] LANG=[$LANG]\n"; read line`
	terminal := `      env: {EDITOR: nano}
      size: {columns: 100, rows: 12}
      actions:
        - wait-for: "COLORTERM=[] NO_COLOR=[] EDITOR=[nano] LANG=[en_US.UTF-8]"`
	if failures := runScenario(t, shellScenario(script, terminal, "")); len(failures) > 0 {
		t.Fatal(failureText(failures))
	}
}
