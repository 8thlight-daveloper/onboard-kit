package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r runner) plan() string { return filepath.Join(r.base, "plan", "plan.obk.yaml") }
func (r runner) accountName() string {
	if r.account == "" {
		return "<account>"
	}
	return r.account
}
func (r runner) preparation() string {
	return filepath.Join(r.base, "preparation", "records", r.accountName(), "preparation.obk.yaml")
}
func (r runner) requireAccount() error {
	if r.account == "" {
		return errors.New("no account is prepared; run just demo-prepare ACCOUNT first")
	}
	return nil
}
func (r runner) preparedAccount() string {
	records := filepath.Join(r.base, "preparation", "records")
	if _, err := r.checked(records, true); err != nil {
		return ""
	}
	entries, err := os.ReadDir(records)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		return ""
	}
	return entries[0].Name()
}
func validAccount(account string) bool {
	return account != "" && account != "." && account != ".." && !strings.ContainsAny(account, `/\`)
}
func (r runner) invitation() string {
	return filepath.Join(r.base, "invitation", "invitation.zip")
}
func (r runner) onboardeeRoot() string {
	return filepath.Join(r.base, "onboardee", "workspace")
}
func (r runner) onboardeeHome() string {
	return filepath.Join(r.base, "onboardee", "home")
}

func (r runner) status() error {
	paths := [][2]string{
		{"Binary", r.binary}, {"Organization plan (autosaved by walk)", r.plan()},
		{"Onboarder preparation", r.preparation()}, {"Invitation", r.invitation()},
		{"Onboardee workspace", r.onboardeeRoot()}, {"Private session registration", r.onboardeeHome()},
		{"Imported preparation", filepath.Join(r.onboardeeRoot(), "records", r.accountName(), "preparation.obk.yaml")},
		{"Onboardee record", filepath.Join(r.onboardeeRoot(), "records", r.accountName(), "record.obk.yaml")},
		{"Derived results only", filepath.Join(r.base, "results")},
		{"Results preparation snapshot", filepath.Join(r.base, "results", "records", r.accountName(), "preparation.obk.yaml")},
		{"Results record snapshot", filepath.Join(r.base, "results", "records", r.accountName(), "record.obk.yaml")},
	}
	key, selectionError := r.sessionKey()
	if selectionError == nil {
		selected := filepath.Join(r.onboardeeHome(), "sessions", key)
		paths = append(paths, [2]string{"Imported organization plan", filepath.Join(selected, "plan.obk.yaml")}, [2]string{"Session metadata", filepath.Join(selected, "session.json")})
	}
	for _, entry := range paths {
		state := "not created"
		if _, err := os.Lstat(entry[1]); err == nil {
			state = "exists"
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if _, err := fmt.Fprintf(r.output, "%s [%s]: %s\n", entry[0], state, entry[1]); err != nil {
			return err
		}
	}
	if selectionError != nil {
		if _, err := fmt.Fprintln(r.output, "Imported plan/session metadata: no session selected (missing, ambiguous, or unsafe registration). No session is guessed."); err != nil {
			return err
		}
	}
	account := "Demo account: " + r.account + "."
	if r.account == "" {
		account = "Demo account: not chosen; just demo-prepare ACCOUNT chooses it."
	}
	_, err := fmt.Fprintln(r.output, account, "Sources are not reset by entering a stage. Use just demo-status in another terminal while a screen is open.")
	return err
}

func (r runner) initializePlan() error {
	if _, err := os.Lstat(r.plan()); err == nil {
		return r.regular(r.plan())
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := r.directory("plan"); err != nil {
		return err
	}
	if err := r.app("new", "--templates", filepath.Join(r.root, "docs", "reference", "examples", "acme", "templates"), "--out", r.plan()); err != nil {
		return err
	}
	if err := r.regular(r.plan()); err != nil {
		return err
	}
	return os.Chmod(r.plan(), 0600)
}

func (r runner) action(action string, args []string) error {
	switch action {
	case "init", "plan":
		if err := r.initializePlan(); err != nil {
			return err
		}
		if action == "init" {
			_, err := fmt.Fprintln(r.output, "New or retained plan:", r.plan())
			return err
		}
		return r.app("walk", "--plan", r.plan(), "--audience", "onboarder")
	case "prepare":
		if err := r.regular(r.plan()); err != nil {
			return fmt.Errorf("complete just demo-plan first: %w", err)
		}
		if existing := r.preparedAccount(); existing != "" && existing != r.account {
			return fmt.Errorf("preparation already exists for %q; run just demo-reset preparation to start another account", existing)
		}
		root := filepath.Join(r.base, "preparation")
		if _, err := os.Lstat(r.preparation()); errors.Is(err, os.ErrNotExist) {
			if err := r.app("prepare", "init", "--plan", r.plan(), "--root", root, "--account", r.account); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := r.app("prepare", "show", "--plan", r.plan(), "--root", root, "--account", r.account); err != nil {
			return err
		}
		_, err := fmt.Fprintf(r.output, "Guided preparation is being rebuilt as a screen. Until then, supply details with the explicit commands, run from %s:\n  %s prepare recipient --plan %s --root %s --account %s --name NAME --onboarding-owner CONTACT\nSee docs/reference/preparation-command.md for every flag.\n", r.root, r.binary, r.plan(), root, r.account)
		return err
	case "invite":
		if err := r.requireAccount(); err != nil {
			return err
		}
		if err := r.regular(r.plan()); err != nil {
			return err
		}
		if err := r.regular(r.preparation()); err != nil {
			return fmt.Errorf("save just demo-prepare first: %w", err)
		}
		return r.app("invite", "export", "--plan", r.plan(), "--preparation", r.preparation(), "--account", r.account, "--out", r.invitation(), "--preaccess-help-contact", args[1])
	case "onboardee":
		if err := r.regular(r.invitation()); err != nil {
			return fmt.Errorf("export just demo-invite CONTACT first: %w", err)
		}
		for _, path := range []string{r.onboardeeRoot(), r.onboardeeHome()} {
			if err := r.mkdir(path); err != nil {
				return err
			}
		}
		return r.app("start", "--invitation", r.invitation(), "--root", r.onboardeeRoot(), "--home", r.onboardeeHome())
	case "progress":
		key, err := r.sessionKey()
		if err != nil {
			return err
		}
		return r.app("session", "edit", "--root", r.onboardeeRoot(), "--home", r.onboardeeHome(), "--session", key)
	case "results":
		return r.results()
	}
	return errors.New("unknown demo action")
}

func (r runner) sessionKey() (string, error) {
	path := filepath.Join(r.onboardeeHome(), "sessions")
	if _, err := r.checked(path, true); err != nil {
		return "", fmt.Errorf("begin just demo-onboardee first: %w", err)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return "", errors.New("expected exactly one demo session; refusing to guess or remove entries")
	}
	name := entries[0].Name()
	decoded, err := hex.DecodeString(name)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("invalid demo session key")
	}
	return name, nil
}

func (r runner) results() error {
	if err := r.requireAccount(); err != nil {
		return err
	}
	if err := r.regular(r.plan()); err != nil {
		return err
	}
	view := filepath.Join(r.base, "results")
	for _, path := range []string{filepath.Join(view, "records"), filepath.Join(view, "records", "newcomer")} {
		if err := r.mkdir(path); err != nil {
			return err
		}
	}
	record := filepath.Join(r.onboardeeRoot(), "records", r.accountName(), "record.obk.yaml")
	for _, pair := range [][2]string{
		{r.preparation(), filepath.Join(view, "records", r.accountName(), "preparation.obk.yaml")},
		{record, filepath.Join(view, "records", r.accountName(), "record.obk.yaml")},
	} {
		if err := r.snapshot(pair[0], pair[1]); err != nil {
			return fmt.Errorf("results need saved preparation and an opened invitation; standard records/ACCOUNT layout required: %w", err)
		}
	}
	if _, err := fmt.Fprintln(r.output, "Results: read-only prepare show JSON, not a graphical dashboard. Copies simulate returning the record; original plan, preparation and record are unchanged."); err != nil {
		return err
	}
	return r.app("prepare", "show", "--plan", r.plan(), "--root", view, "--account", r.account)
}
