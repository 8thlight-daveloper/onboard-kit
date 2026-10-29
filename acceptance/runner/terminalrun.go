package runner

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
)

const terminalWaitLimit = 5 * time.Second

var trailingNumber = regexp.MustCompile(`(\d+)\D*$`)

var (
	collectorLock    sync.Mutex
	collectorReaders int
	collectorPercent int
)

type terminalRun struct {
	client  *tuitest.Client
	process int
	limit   time.Time
}

func runTerminal(step Step, folder string, options Options) (outcome, []Failure) {
	client, err := tuitest.Ephemeral("scenario", tuitest.ClientOptions{})
	if err != nil {
		return outcome{}, []Failure{{Expectation: "start", Message: err.Error()}}
	}
	defer client.CloseQuiet()
	program := step.Run[0]
	if program == binaryPlaceholder {
		program = options.BinaryPath
	}
	columns, rows := uint16(step.Terminal.Size.Columns), uint16(step.Terminal.Size.Rows)
	started, err := client.Run(program, step.Run[1:], tuitest.SpawnOptions{
		Cols: &columns, Rows: &rows, Cwd: folder, Env: terminalEnvironment(step.Terminal, folder), WaitReady: tuitest.Ptr(false),
	})
	if err != nil {
		return outcome{}, []Failure{{Expectation: "start", Message: err.Error()}}
	}
	run := &terminalRun{client: client, limit: time.Now().Add(options.timeLimit())}
	if started.ShellPID != nil {
		run.process = int(*started.ShellPID)
	}
	defer run.stop()
	for position, action := range step.Terminal.Actions {
		if failure := run.perform(action); failure != nil {
			failure.Expectation = fmt.Sprintf("terminal action %d, %s", position+1, failure.Expectation)
			return outcome{}, []Failure{*failure, run.screenFailure()}
		}
	}
	result := outcome{}
	if step.Expect.Exit != nil {
		code, failure := run.exitCode()
		if failure != nil {
			return outcome{}, []Failure{*failure, run.screenFailure()}
		}
		result.ExitCode = code
	}
	return result, nil
}

func terminalEnvironment(terminal *Terminal, folder string) map[string]string {
	environment := map[string]string{"PATH": os.Getenv("PATH"), "HOME": folder, "TERM": "xterm-256color"}
	for name, value := range terminal.Env {
		environment[name] = value
	}
	return environment
}

func (run *terminalRun) stop() {
	if run.process > 0 {
		syscall.Kill(run.process, syscall.SIGKILL)
	}
}

func (run *terminalRun) screenFailure() Failure {
	text, err := run.client.Text(tuitest.TextOptions{})
	if err != nil {
		text = err.Error()
	}
	return Failure{Expectation: "screen", Message: "the screen at the failure:\n" + text}
}

func (run *terminalRun) remaining() time.Duration {
	return min(terminalWaitLimit, time.Until(run.limit))
}

func (run *terminalRun) perform(action TerminalAction) *Failure {
	switch {
	case action.WaitFor != nil:
		return run.waitForText(*action.WaitFor, tuitest.Visible)
	case action.WaitGone != nil:
		return run.waitForText(*action.WaitGone, tuitest.Hidden)
	case action.Press != nil:
		return usage("press", run.client.Press(action.Press...))
	case action.Type != nil:
		return usage("type", run.client.Type(*action.Type))
	case action.Resize != nil:
		return usage("resize", run.client.Resize(uint16(action.Resize.Columns), uint16(action.Resize.Rows)))
	case action.Signal != nil:
		return run.send(strings.ToUpper(*action.Signal))
	case action.WaitExit:
		_, failure := run.exitCode()
		return failure
	case action.ExpectModes != nil:
		return run.expectModes(action.ExpectModes)
	default:
		return run.expectTerminalColors()
	}
}

func usage(name string, err error) *Failure {
	if err == nil {
		return nil
	}
	return &Failure{Expectation: name, Message: err.Error()}
}

func (run *terminalRun) waitForText(text string, visibility tuitest.Visibility) *Failure {
	limit := run.remaining()
	if limit <= 0 {
		return &Failure{Expectation: "time limit", Message: "the step ran out of time"}
	}
	err := run.client.GetByText(text, tuitest.TextSelectorOptions{}).Wait(tuitest.LocatorWaitOptions{State: visibility, Timeout: &limit})
	if err == nil {
		return nil
	}
	name := "wait-for " + strconv.Quote(text)
	if visibility == tuitest.Hidden {
		name = "wait-gone " + strconv.Quote(text)
	}
	return &Failure{Expectation: name, Message: err.Error()}
}

func (run *terminalRun) send(name string) *Failure {
	if run.process <= 0 {
		return &Failure{Expectation: "signal", Message: "the process id is not known"}
	}
	if err := syscall.Kill(run.process, terminalSignals[name]); err != nil {
		return &Failure{Expectation: "signal " + name, Message: err.Error()}
	}
	return nil
}

func (run *terminalRun) exitCode() (int, *Failure) {
	for time.Now().Before(run.limit) && run.remaining() > 0 {
		state, err := run.client.State()
		if err != nil {
			return 0, &Failure{Expectation: "exit", Message: err.Error()}
		}
		if state.Exited != nil {
			return exitStatus(state), nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return 0, &Failure{Expectation: "exit", Message: "the process did not exit in time"}
}

func exitStatus(state tuitest.State) int {
	if state.ExitSignal != nil {
		if match := trailingNumber.FindStringSubmatch(*state.ExitSignal); match != nil {
			number, _ := strconv.Atoi(match[1])
			return 128 + number
		}
	}
	return int(*state.Exited)
}

func (run *terminalRun) expectModes(expected map[string]bool) *Failure {
	state, err := run.client.State()
	if err != nil {
		return &Failure{Expectation: "expect-modes", Message: err.Error()}
	}
	var wrong []string
	for _, name := range sortedModeNames(expected) {
		if state.Modes[name] != expected[name] {
			wrong = append(wrong, fmt.Sprintf("%s expected %t, got %t", name, expected[name], state.Modes[name]))
		}
	}
	if len(wrong) == 0 {
		return nil
	}
	return &Failure{Expectation: "expect-modes", Message: strings.Join(wrong, "; ")}
}

func sortedModeNames(modes map[string]bool) []string {
	names := make([]string, 0, len(modes))
	for name := range modes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (run *terminalRun) expectTerminalColors() *Failure {
	state, err := run.client.State()
	if err != nil {
		return &Failure{Expectation: "expect-colors", Message: err.Error()}
	}
	var cells []tuitest.Cell
	withoutGarbageCollection(func() {
		cells, err = run.client.Cells(0, 0, state.Cols, state.Rows)
	})
	if err != nil {
		return &Failure{Expectation: "expect-colors", Message: err.Error()}
	}
	var offenders []string
	for _, cell := range cells {
		for _, color := range []tuitest.Color{cell.FG, cell.BG} {
			if !terminalColor(color) && len(offenders) < 5 {
				offenders = append(offenders, fmt.Sprintf("column %d, row %d, %q is %q", cell.X, cell.Y, cell.Char, color))
			}
		}
	}
	if len(offenders) == 0 {
		return nil
	}
	return &Failure{Expectation: "expect-colors", Message: "a cell is neither the default color nor one of the terminal's sixteen: " + strings.Join(offenders, "; ")}
}

func terminalColor(color tuitest.Color) bool {
	if color == "default" {
		return true
	}
	index, err := strconv.Atoi(string(color))
	return err == nil && index >= 0 && index <= 15
}

func withoutGarbageCollection(read func()) {
	collectorLock.Lock()
	if collectorReaders == 0 {
		collectorPercent = debug.SetGCPercent(-1)
		runtime.GC()
	}
	collectorReaders++
	collectorLock.Unlock()
	read()
	collectorLock.Lock()
	collectorReaders--
	if collectorReaders == 0 {
		debug.SetGCPercent(collectorPercent)
	}
	collectorLock.Unlock()
}
