package runner

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"go.yaml.in/yaml/v3"
)

const (
	colorsTerminalOnly = "terminal-only"
	colorsPlain        = "plain"

	defaultTerminalColumns = 80
	defaultTerminalRows    = 24
)

var terminalSignals = map[string]syscall.Signal{
	"INT":  syscall.SIGINT,
	"TERM": syscall.SIGTERM,
	"HUP":  syscall.SIGHUP,
	"QUIT": syscall.SIGQUIT,
}

var terminalModes = map[string]bool{
	"alternate_screen": true,
	"cursor_visible":   true,
	"bracketed_paste":  true,
}

type Terminal struct {
	Size    TerminalSize      `yaml:"size"`
	Env     map[string]string `yaml:"env"`
	Actions []TerminalAction  `yaml:"actions"`
}

type TerminalSize struct {
	Columns int
	Rows    int
}

type TerminalAction struct {
	WaitFor      *string
	WaitGone     *string
	Press        []string
	Type         *string
	Resize       *TerminalSize
	Signal       *string
	WaitExit     bool
	ExpectModes  map[string]bool
	ExpectColors *string
}

func (terminal *Terminal) UnmarshalYAML(node *yaml.Node) error {
	err := decodeFields(node, map[string]func(*yaml.Node) error{
		"size":    func(value *yaml.Node) error { return value.Decode(&terminal.Size) },
		"env":     func(value *yaml.Node) error { return value.Decode(&terminal.Env) },
		"actions": func(value *yaml.Node) error { return value.Decode(&terminal.Actions) },
	})
	if err != nil {
		return err
	}
	if terminal.Size.Columns == 0 && terminal.Size.Rows == 0 {
		terminal.Size = TerminalSize{Columns: defaultTerminalColumns, Rows: defaultTerminalRows}
	}
	return nil
}

func (size *TerminalSize) UnmarshalYAML(node *yaml.Node) error {
	return decodeFields(node, map[string]func(*yaml.Node) error{
		"columns": func(value *yaml.Node) error { return value.Decode(&size.Columns) },
		"rows":    func(value *yaml.Node) error { return value.Decode(&size.Rows) },
	})
}

func (action *TerminalAction) UnmarshalYAML(node *yaml.Node) error {
	held := 0
	hold := func(decode func(*yaml.Node) error) func(*yaml.Node) error {
		return func(value *yaml.Node) error {
			held++
			return decode(value)
		}
	}
	err := decodeFields(node, map[string]func(*yaml.Node) error{
		"wait-for":  hold(func(value *yaml.Node) error { return value.Decode(&action.WaitFor) }),
		"wait-gone": hold(func(value *yaml.Node) error { return value.Decode(&action.WaitGone) }),
		"press": hold(func(value *yaml.Node) error {
			if value.Kind == yaml.ScalarNode {
				action.Press = []string{value.Value}
				return nil
			}
			return value.Decode(&action.Press)
		}),
		"type":          hold(func(value *yaml.Node) error { return value.Decode(&action.Type) }),
		"resize":        hold(func(value *yaml.Node) error { return value.Decode(&action.Resize) }),
		"signal":        hold(func(value *yaml.Node) error { return value.Decode(&action.Signal) }),
		"wait-exit":     hold(func(value *yaml.Node) error { return value.Decode(&action.WaitExit) }),
		"expect-modes":  hold(func(value *yaml.Node) error { return value.Decode(&action.ExpectModes) }),
		"expect-colors": hold(func(value *yaml.Node) error { return value.Decode(&action.ExpectColors) }),
	})
	if err != nil {
		return err
	}
	if held != 1 {
		return fmt.Errorf("line %d: an action has exactly one field", node.Line)
	}
	return nil
}

func validateTerminal(index int, step Step) error {
	terminal := step.Terminal
	label := fmt.Sprintf("step %d", index+1)
	switch {
	case step.Talk != nil:
		return fmt.Errorf("%s: a terminal step cannot also have talk", label)
	case step.Stdin != "":
		return fmt.Errorf("%s: a terminal step cannot also have stdin", label)
	case step.Capture != nil:
		return fmt.Errorf("%s: a terminal step cannot also have capture", label)
	case step.Expect.Stdout != nil || step.Expect.Stderr != nil:
		return fmt.Errorf("%s: a terminal step has no stdout or stderr to check", label)
	case terminal.Size.Columns < 1 || terminal.Size.Rows < 1:
		return fmt.Errorf("%s: the terminal size needs columns and rows of at least 1", label)
	}
	for position, action := range terminal.Actions {
		if err := validateAction(action); err != nil {
			return fmt.Errorf("%s, action %d: %w", label, position+1, err)
		}
	}
	return nil
}

func validateAction(action TerminalAction) error {
	if action.Signal != nil {
		if _, known := terminalSignals[strings.ToUpper(*action.Signal)]; !known {
			return fmt.Errorf("signal %q is not INT, TERM, HUP, or QUIT", *action.Signal)
		}
	}
	for mode := range action.ExpectModes {
		if !terminalModes[mode] {
			return fmt.Errorf("mode %q is not alternate_screen, cursor_visible, or bracketed_paste", mode)
		}
	}
	if action.ExpectColors != nil && *action.ExpectColors != colorsTerminalOnly && *action.ExpectColors != colorsPlain {
		return errors.New(`expect-colors takes "terminal-only" or "plain"`)
	}
	if action.Resize != nil && (action.Resize.Columns < 1 || action.Resize.Rows < 1) {
		return errors.New("resize needs columns and rows of at least 1")
	}
	if len(action.Press) == 1 && action.Press[0] == "" {
		return errors.New("press needs a key")
	}
	return nil
}
