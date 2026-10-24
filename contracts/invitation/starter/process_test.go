package starter_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	contract "onboard-kit/contracts/invitation/starter"
	"onboard-kit/internal/model"
)

func TestStarterNativeChild(t *testing.T) {
	if os.Getenv("OBK_STARTER_CONTRACT_CHILD") != "1" {
		return
	}
	var args []string
	if json.Unmarshal([]byte(os.Getenv("OBK_STARTER_CONTRACT_ARGS")), &args) != nil {
		os.Exit(97)
	}
	command := newStarter(contract.Runtime{Clock: func() model.Timestamp { return openedAt }})
	os.Exit(command.Run(args, os.Stdin, os.Stdout, os.Stderr))
}

type child struct {
	command *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Reader
	errors  bytes.Buffer
	cancel  context.CancelFunc
	waited  bool
}

func launchChild(t *testing.T, s sandbox) *child {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	c := &child{command: exec.CommandContext(ctx, executable, "-test.run=^TestStarterNativeChild$"), cancel: cancel}
	args, err := json.Marshal(s.args())
	if err != nil {
		t.Fatal(err)
	}
	c.command.Env = append(os.Environ(), "OBK_STARTER_CONTRACT_CHILD=1", "OBK_STARTER_CONTRACT_ARGS="+string(args))
	c.command.Stderr = &c.errors
	c.input, err = c.command.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	out, err := c.command.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	c.output = bufio.NewReader(out)
	if err = c.command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !c.waited {
			_ = c.input.Close()
			c.cancel()
			_ = c.command.Wait()
			c.waited = true
		}
		c.cancel()
	})
	return c
}

func (c *child) until(t *testing.T, prompt string) {
	t.Helper()
	total := 0
	for {
		line, err := c.output.ReadString('\n')
		total += len(line)
		if total > 1048576 {
			t.Fatal("bounded native helper output exceeded")
		}
		if err != nil {
			t.Fatal("native helper ended before required public prompt")
		}
		if line == prompt+"\n" {
			return
		}
	}
}

func (c *child) send(t *testing.T, text string) {
	t.Helper()
	if _, err := io.WriteString(c.input, text); err != nil {
		t.Fatal("cannot send owned native helper input")
	}
}
func (c *child) finish(t *testing.T) {
	t.Helper()
	_ = c.input.Close()
	_, _ = io.Copy(io.Discard, c.output)
	err := c.command.Wait()
	c.waited = true
	c.cancel()
	if err != nil {
		t.Fatal("owned native helper did not finish successfully")
	}
}

func TestActualTwoProcessExclusionAndSessionRelease(t *testing.T) {
	s := setup(t)
	c := launchChild(t, s)
	c.until(t, "Identity: choose confirm jules or cancel:")
	before := snapshot(t, s.base)
	clock := 0
	blocked := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\n")
	wantReport(t, blocked, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
	if clock != 0 {
		t.Error("native exclusion loser sampled event")
	}
	unchanged(t, s, before)
	c.send(t, "confirm jules\n")
	c.until(t, "Review complete. Choose begin or cancel:")
	c.send(t, "begin\n")
	c.until(t, "Session: read only. Choose exit:")
	second := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "cancel\n")
	wantReport(t, second, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
	if clock != 0 {
		t.Error("independent identity sampled an event")
	}
	c.send(t, "exit\n")
	c.finish(t)
	if !strings.Contains(c.errors.String(), `"code":"start.opened"`) {
		t.Error("native helper opening did not complete")
	}
	assertBytes(t, s.record(), fixture(t, "begun.yaml"))
	assertIdle(t, s)
}
