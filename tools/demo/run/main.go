package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var stages = []string{"plan", "preparation", "invitation", "onboardee", "results"}

type runner struct {
	root, base, binary, account string
	input                       io.Reader
	output, errors              io.Writer
}

func main() {
	root, err := os.Getwd()
	if err == nil {
		err = execute(root, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func execute(root string, args []string, input io.Reader, output, errorOutput io.Writer) error {
	if len(args) == 0 {
		return errors.New("choose init, status, plan, prepare ACCOUNT, invite CONTACT, onboardee, progress, results, or reset STAGE")
	}
	action := args[0]
	count := 1
	if action == "prepare" || action == "invite" || action == "reset" {
		count = 2
	}
	if len(args) != count {
		return errors.New("incorrect demo arguments; prepare requires ACCOUNT, invite requires CONTACT and reset requires STAGE")
	}
	stage := map[string]string{"init": "plan", "plan": "plan", "prepare": "preparation", "invite": "invitation", "onboardee": "onboardee", "progress": "onboardee", "results": "results"}[action]
	if action == "reset" {
		stage = args[1]
		if !knownStage(stage) && stage != "all" {
			return errors.New("reset stage must be plan, preparation, invitation, onboardee, results, or all")
		}
	} else if action != "status" && stage == "" {
		return errors.New("unknown demo action")
	}
	if action == "prepare" && args[1] == "" {
		return errors.New("say whose preparation this is, for example: just demo-prepare jules")
	}
	if action == "prepare" && !validAccount(args[1]) {
		return errors.New("the account is used as a folder name, so use one word with no slash")
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(physical)
	if err != nil {
		return err
	}
	r := runner{root: absolute, base: filepath.Join(absolute, "output", "manual-test"), binary: filepath.Join(absolute, "output", "demo-bin", "onboard-kit"), input: input, output: output, errors: errorOutput}
	r.account = r.preparedAccount()
	if action == "prepare" {
		r.account = args[1]
	}
	if action == "status" {
		return r.status()
	}
	if err := r.workspace(); err != nil {
		return err
	}
	selected := []string{stage}
	if stage == "all" {
		selected = stages
	}
	for _, name := range selected {
		path := filepath.Join(r.base, name)
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		if _, err := r.checked(path, true); err != nil {
			return err
		}
	}
	var locks []*os.File
	defer func() {
		for _, lock := range locks {
			lock.Close()
		}
	}()
	for _, name := range selected {
		lock, err := r.lock(name)
		if err != nil {
			return err
		}
		locks = append(locks, lock)
	}
	if action == "reset" {
		for _, name := range selected {
			if err := r.reset(name); err != nil {
				return err
			}
		}
		return nil
	}
	if err := r.status(); err != nil {
		return err
	}
	if err := r.directory(stage); err != nil {
		return err
	}
	return r.action(action, args)
}

func knownStage(stage string) bool {
	for _, name := range stages {
		if name == stage {
			return true
		}
	}
	return false
}

func (r runner) app(args ...string) error {
	if err := r.regular(r.binary); err != nil {
		return fmt.Errorf("build the demo binary with just demo-build: %w", err)
	}
	command := exec.Command(r.binary, args...)
	command.Dir = r.root
	command.Stdin, command.Stdout, command.Stderr = r.input, r.output, r.errors
	return command.Run()
}
