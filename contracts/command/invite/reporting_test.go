package invite_test

import (
	"bytes"
	"errors"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("PRIVATE_WRITER_ERROR") }

func TestStdoutFailureAfterPublicationKeepsCompleteArchive(t *testing.T) {
	s := newSandbox(t)
	before := tree(t, s.root)
	var stderr bytes.Buffer
	exit := newCommand().Run(s.args(), failingWriter{}, &stderr)
	readReport(t, result{exit: exit, stderr: stderr.String()}, "invite.result-output", 1, nil, nil)
	preserved(t, s, before, true)
	exportedPayloads(t, s.out)
}

func TestStderrFailureAfterPublicationIsNotSilentSuccess(t *testing.T) {
	s := newSandbox(t)
	before := tree(t, s.root)
	var stdout bytes.Buffer
	exit := newCommand().Run(s.args(), &stdout, failingWriter{})
	if exit != 1 || stdout.String() != success {
		t.Fatal("report failure did not follow specified postcommit ordering")
	}
	preserved(t, s, before, true)
	exportedPayloads(t, s.out)
}

func TestHelpOutputFailureHasNoFileEffects(t *testing.T) {
	s := newSandbox(t)
	before := tree(t, s.root)
	var stderr bytes.Buffer
	exit := newCommand().Run([]string{"--help"}, failingWriter{}, &stderr)
	readReport(t, result{exit: exit, stderr: stderr.String()}, "invite.help-output", 1, nil, nil)
	preserved(t, s, before, false)
}
