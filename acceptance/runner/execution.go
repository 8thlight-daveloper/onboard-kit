package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const binaryPlaceholder = "onboard-kit"

type outcome struct {
	Stdout        string
	Stderr        string
	ExitCode      int
	TimedOut      bool
	TalkFailures  []Failure
	ProcessFailed error
}

func command(step Step, folder string, options Options, limit context.Context) *exec.Cmd {
	program := step.Run[0]
	if program == binaryPlaceholder {
		program = options.BinaryPath
	}
	process := exec.CommandContext(limit, program, step.Run[1:]...)
	process.Dir = folder
	process.WaitDelay = time.Second
	return process
}

func runCommand(step Step, folder string, options Options) outcome {
	limit, cancel := context.WithTimeout(context.Background(), options.timeLimit())
	defer cancel()
	process := command(step, folder, options, limit)
	var stdout, stderr bytes.Buffer
	process.Stdin = strings.NewReader(step.Stdin)
	process.Stdout = &stdout
	process.Stderr = &stderr
	err := process.Run()
	return finish(process, err, limit, stdout.String(), stderr.String())
}

func finish(process *exec.Cmd, err error, limit context.Context, stdout, stderr string) outcome {
	result := outcome{Stdout: stdout, Stderr: stderr}
	if errors.Is(limit.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		return result
	}
	if process.ProcessState == nil {
		result.ProcessFailed = err
		return result
	}
	result.ExitCode = process.ProcessState.ExitCode()
	return result
}

type line struct {
	Text string
	Err  error
}

func runStdio(step Step, folder string, options Options) outcome {
	limit, cancel := context.WithTimeout(context.Background(), options.timeLimit())
	defer cancel()
	process := command(step, folder, options, limit)
	var stderr bytes.Buffer
	process.Stderr = &stderr
	input, err := process.StdinPipe()
	if err != nil {
		return outcome{ProcessFailed: err}
	}
	output, err := process.StdoutPipe()
	if err != nil {
		return outcome{ProcessFailed: err}
	}
	if err := process.Start(); err != nil {
		return outcome{ProcessFailed: err}
	}
	lines := readLines(bufio.NewReader(output))
	var stdout strings.Builder
	failures, timedOut := talk(step.Talk, input, lines, limit, &stdout)
	input.Close()
	for remaining := range lines {
		stdout.WriteString(remaining.Text)
	}
	waitErr := process.Wait()
	result := finish(process, waitErr, limit, stdout.String(), stderr.String())
	result.TalkFailures = failures
	result.TimedOut = result.TimedOut || timedOut
	return result
}

func readLines(reader *bufio.Reader) <-chan line {
	lines := make(chan line)
	go func() {
		defer close(lines)
		for {
			text, err := reader.ReadString('\n')
			if text != "" {
				lines <- line{Text: text}
			}
			if err != nil {
				return
			}
		}
	}()
	return lines
}

func talk(exchanges []Exchange, input io.Writer, lines <-chan line, limit context.Context, stdout *strings.Builder) ([]Failure, bool) {
	for index, exchange := range exchanges {
		label := fmt.Sprintf("talk %d", index+1)
		request, err := json.Marshal(exchange.Send)
		if err != nil {
			return []Failure{{Expectation: label, Message: fmt.Sprintf("the request cannot be written as JSON: %v", err)}}, false
		}
		if _, err := input.Write(append(request, '\n')); err != nil {
			return []Failure{{Expectation: label, Message: fmt.Sprintf("the request could not be sent: %v", err)}}, false
		}
		select {
		case response, open := <-lines:
			if !open {
				return []Failure{{Expectation: label, Message: "the process closed its output before it answered"}}, false
			}
			stdout.WriteString(response.Text)
			if failure := checkResponse(label, exchange, response.Text); failure != nil {
				return []Failure{*failure}, false
			}
		case <-limit.Done():
			return nil, true
		}
	}
	return nil, false
}

func checkResponse(label string, exchange Exchange, text string) *Failure {
	if exchange.Expect == nil {
		return nil
	}
	var actual any
	if err := json.Unmarshal([]byte(text), &actual); err != nil {
		return &Failure{Expectation: label, Message: fmt.Sprintf("the response is not JSON: %q", text)}
	}
	if mismatch := subsetMismatch(exchange.Expect.Value, actual, "$"); mismatch != "" {
		return &Failure{Expectation: label, Message: fmt.Sprintf("%s in response %q", mismatch, strings.TrimSpace(text))}
	}
	return nil
}
