package invite_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFilesystemRefusalsPreserveEveryObject(t *testing.T) {
	for _, name := range []string{"output-file", "output-directory", "output-symlink", "dangling-output-symlink", "output-hardlink", "output-parent-symlink", "missing-output-parent", "output-dotdot", "plan-symlink", "plan-parent-symlink", "preparation-symlink", "plan-directory", "plan-fifo", "plan-too-large", "preparation-too-large"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t)
			code := "invite.output-exists"
			switch name {
			case "output-file":
				writeLiteral(t, s.out, []byte("PRIVATE_EXISTING_ARCHIVE\n"))
			case "output-directory":
				must(t, os.Mkdir(s.out, 0o700))
			case "output-symlink", "dangling-output-symlink":
				target := filepath.Join(s.root, "target")
				if name == "output-symlink" {
					writeLiteral(t, target, []byte("PRIVATE_TARGET\n"))
				}
				must(t, os.Symlink(target, s.out))
			case "output-hardlink":
				must(t, os.Link(s.plan, s.out))
			case "output-parent-symlink":
				must(t, os.Mkdir(filepath.Join(s.root, "real"), 0o700))
				must(t, os.Symlink(filepath.Join(s.root, "real"), filepath.Join(s.root, "alias")))
				s.out = filepath.Join(s.root, "alias", "out.zip")
				code = "invite.output-path"
			case "missing-output-parent":
				s.out = filepath.Join(s.root, "absent", "out.zip")
				code = "invite.output-path"
			case "output-dotdot":
				s.out = s.root + "/../" + filepath.Base(s.root) + "/out.zip"
				code = "invite.output-path"
			case "plan-symlink", "preparation-symlink":
				path := s.plan
				code = "invite.plan-read"
				if name == "preparation-symlink" {
					path, code = s.preparation, "invite.preparation-read"
				}
				must(t, os.Rename(path, path+".real"))
				must(t, os.Symlink(path+".real", path))
			case "plan-parent-symlink":
				must(t, os.Symlink(s.root, filepath.Join(s.root, "alias")))
				s.plan = filepath.Join(s.root, "alias", "plan.yaml")
				code = "invite.plan-read"
			case "plan-directory", "plan-fifo":
				must(t, os.Remove(s.plan))
				if name == "plan-directory" {
					must(t, os.Mkdir(s.plan, 0o700))
				} else {
					must(t, syscall.Mkfifo(s.plan, 0o600))
				}
				code = "invite.plan-read"
			case "plan-too-large":
				writeLiteral(t, s.plan, bytes.Repeat([]byte{'x'}, 1048577))
				code = "invite.plan-read"
			case "preparation-too-large":
				writeLiteral(t, s.preparation, bytes.Repeat([]byte{'x'}, 1048577))
				code = "invite.preparation-read"
			}
			before := tree(t, s.root)
			got := run(newCommand(), s.args())
			readReport(t, got, code, 1, nil, nil)
			if strings.Contains(got.stdout+got.stderr, s.root) {
				t.Fatal("diagnostic exposed the raw selected path")
			}
			preserved(t, s, before, false)
		})
	}
}

func TestInputBudgetAcceptsExactlyOneMiBOfRealPlan(t *testing.T) {
	s := newSandbox(t)
	base := literal(t, "plan.yaml")
	padding := "#" + strings.Repeat("x", 1048576-len(base)-2) + "\n"
	writeLiteral(t, s.plan, append(base, []byte(padding)...))
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	if !bytes.Equal(payloads["plan.obk.yaml"], literal(t, "plan.yaml")) {
		t.Fatal("raw source padding changed canonical plan payload")
	}
}

func TestActualCanonicalPayloadAggregateBoundary(t *testing.T) {
	baseTotal := len(literal(t, "plan.yaml")) + len(literal(t, "preparation.yaml")) + len(literal(t, "record.yaml")) + len(literal(t, "first-steps.md"))
	for _, excess := range []int{0, 1} {
		t.Run(string(rune('0'+excess)), func(t *testing.T) {
			s := newSandbox(t)
			paddingSize := 1048576 - baseTotal - len("padding: \n") + excess
			if paddingSize < 1 {
				t.Fatal("literal baseline exceeds practical boundary fixture")
			}
			preparation := append(literal(t, "preparation.yaml"), []byte("padding: "+strings.Repeat("x", paddingSize)+"\n")...)
			if len(preparation) > 1048576 {
				t.Fatal("boundary fixture accidentally hits raw-input cap first")
			}
			writeLiteral(t, s.preparation, preparation)
			before := tree(t, s.root)
			got := run(newCommand(), s.args())
			if excess == 1 {
				readReport(t, got, "invite.archive", 1, []projectedProblem{{Code: "invitation.too-large", Severity: "error"}}, nil)
				preserved(t, s, before, false)
				return
			}
			readReport(t, got, "invite.written", 0, nil, nil)
			preserved(t, s, before, true)
			payloads := exportedPayloads(t, s.out)
			total := 0
			for _, payload := range payloads {
				total += len(payload)
			}
			if total != 1048576 || !bytes.Equal(payloads["records/newcomer/preparation.obk.yaml"], preparation) {
				t.Fatal("boundary did not exercise exact actual canonical payload sum")
			}
		})
	}
}

func TestTwoRealCommandsCompeteWithoutReplacingOutput(t *testing.T) {
	s := newSandbox(t)
	before := tree(t, s.root)
	start := make(chan struct{})
	var wait sync.WaitGroup
	results := make([]result, 2)
	for index := range results {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results[index] = run(newCommand(), s.args())
		}()
	}
	close(start)
	wait.Wait()
	successes := 0
	for _, got := range results {
		if got.exit == 0 {
			readReport(t, got, "invite.written", 0, nil, nil)
			successes++
		} else {
			readReport(t, got, "invite.output-exists", 1, nil, nil)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one complete publication, got %d", successes)
	}
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	if !bytes.Equal(payloads["plan.obk.yaml"], literal(t, "plan.yaml")) || !bytes.Equal(payloads["first-steps.md"], literal(t, "first-steps.md")) {
		t.Fatal("race published altered/mixed archive")
	}
}
