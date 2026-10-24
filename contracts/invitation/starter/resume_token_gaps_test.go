package starter_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type resumeTokenUnreadInput struct {
	source *strings.Reader
	reads  int
}

func (r *resumeTokenUnreadInput) Read(p []byte) (int, error) {
	r.reads++
	return r.source.Read(p)
}

func TestValidResumeTokenIsKnownBeforeContextVerification(t *testing.T) {
	for _, context := range []string{"absent-home", "absent-namespace", "missing-root", "symlink-root", "non-directory-root"} {
		for _, equals := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/equals=%t", context, equals), func(t *testing.T) {
				s := setup(t)
				root := s.root
				switch context {
				case "absent-home":
				case "absent-namespace":
					if err := os.Mkdir(s.home, 0700); err != nil {
						t.Fatal(err)
					}
				case "missing-root":
					root = filepath.Join(s.base, "missing-root")
				case "symlink-root":
					root = filepath.Join(s.base, "root-alias")
					if err := os.Symlink(s.root, root); err != nil {
						t.Fatal(err)
					}
				case "non-directory-root":
					root = filepath.Join(s.base, "root-file")
					put(t, root, []byte("OWNED CONTEXT FILE\n"))
				}
				token := strings.Repeat("0123456789abcdef", 4)
				args := []string{"--resume-import", token, "--root", root, "--home", s.home}
				if equals {
					token = strings.Repeat("fedcba9876543210", 4)
					args = []string{"--home=" + s.home, "--root=" + root, "--resume-import=" + token}
				}
				before := snapshot(t, s.base)
				clock := 0
				input := &resumeTokenUnreadInput{source: strings.NewReader("resume\nclean\nexit\nUNUSED\n")}
				remaining := input.source.Len()
				var stdout, stderr bytes.Buffer
				exit := newStarter(runtimeFacts(t, &clock)).Run(args, input, &stdout, &stderr)
				want := fmt.Sprintf("{\"status\":\"refused\",\"code\":\"start.context\",\"message\":\"Select a safe supported root, state home, and unchanged session context.\",\"transaction\":\"%s\"}\n", token)
				if exit != 1 || stdout.Len() != 0 || stderr.String() != want {
					t.Errorf("valid resume context refusal must retain the exact supplied token: exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
				}
				if clock != 0 || input.reads != 0 || input.source.Len() != remaining {
					t.Errorf("context refusal must read no action or clock: clock=%d reads=%d remaining=%d", clock, input.reads, input.source.Len())
				}
				unchanged(t, s, before)
			})
		}
	}
}

func TestResumeTokenRemainsUnknownUntilTheWholeCLIIsValid(t *testing.T) {
	token := strings.Repeat("0123456789abcdef", 4)
	for _, kind := range []string{"missing-root-option", "missing-token-value", "empty-token", "short-token", "long-token", "uppercase-token", "unknown-after-token", "unknown-before-token", "duplicate-token", "mixed-mode", "late-help", "positional-tail", "empty-home"} {
		t.Run(kind, func(t *testing.T) {
			s := setup(t)
			args := []string{"--resume-import=" + token, "--root=" + s.root, "--home=" + s.home}
			switch kind {
			case "missing-root-option":
				args = []string{"--resume-import=" + token, "--home=" + s.home}
			case "missing-token-value":
				args = []string{"--root=" + s.root, "--home=" + s.home, "--resume-import"}
			case "empty-token":
				args[0] = "--resume-import="
			case "short-token":
				args[0] = "--resume-import=" + token[:63]
			case "long-token":
				args[0] = "--resume-import=" + token + "0"
			case "uppercase-token":
				args[0] = "--resume-import=" + strings.ToUpper(token)
			case "unknown-after-token":
				args = append(args, "--unknown")
			case "unknown-before-token":
				args = append([]string{"--unknown"}, args...)
			case "duplicate-token":
				args = append(args, "--resume-import="+token)
			case "mixed-mode":
				args = append(args, "--invitation="+s.invitation)
			case "late-help":
				args = append(args, "--help")
			case "positional-tail":
				args = append(args, "not-an-option")
			case "empty-home":
				args[2] = "--home="
			}
			before := snapshot(t, s.base)
			clock := 0
			input := &resumeTokenUnreadInput{source: strings.NewReader("resume\nUNUSED\n")}
			remaining := input.source.Len()
			var stdout, stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(args, input, &stdout, &stderr)
			want := "{\"status\":\"refused\",\"code\":\"start.usage\",\"message\":\"Use one invitation or resume token, an explicit root, and an optional home.\",\"transaction\":\"\"}\n"
			if exit != 2 || stdout.Len() != 0 || stderr.String() != want {
				t.Errorf("any invalid CLI must retain an empty transaction: exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
			}
			if clock != 0 || input.reads != 0 || input.source.Len() != remaining {
				t.Errorf("invalid CLI must read no action or clock: clock=%d reads=%d remaining=%d", clock, input.reads, input.source.Len())
			}
			unchanged(t, s, before)
		})
	}
}

func TestInitialHelpDoesNotReportASuppliedResumeToken(t *testing.T) {
	for _, help := range []string{"-h", "--help"} {
		t.Run(help, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			input := &resumeTokenUnreadInput{source: strings.NewReader("resume\nUNUSED\n")}
			remaining := input.source.Len()
			args := []string{help, "--resume-import=" + strings.Repeat("0123456789abcdef", 4), "--root=" + s.root, "--home=" + s.home, "--unknown"}
			var stdout, stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(args, input, &stdout, &stderr)
			if exit != 0 || !bytes.Equal(stdout.Bytes(), fixture(t, "help.txt")) || stderr.Len() != 0 {
				t.Error("initial help must retain exact help bytes and empty stderr")
			}
			if clock != 0 || input.reads != 0 || input.source.Len() != remaining {
				t.Error("initial help must read no action or clock")
			}
			unchanged(t, s, before)
		})
	}
}

var _ io.Reader = (*resumeTokenUnreadInput)(nil)
