package creator_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type finalFaultReader struct {
	source   []byte
	position int
	final    error
}

func (r *finalFaultReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.position == len(r.source) {
		return 0, io.EOF
	}
	p[0] = r.source[r.position]
	r.position++
	if r.position == len(r.source) {
		return 1, r.final
	}
	return 1, nil
}

func TestPositiveFinalReadDistinguishesEOFAndOtherError(t *testing.T) {
	for _, fail := range []bool{false, true} {
		root, target := sandbox(t)
		var out, stderr bytes.Buffer
		terminal := io.EOF
		if fail {
			terminal = errors.New("owned input failure")
		}
		input := &finalFaultReader{source: []byte("review\nsave\n"), final: terminal}
		exit := newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr)
		if fail {
			if exit != 1 {
				t.Error("nonEOF error paired with save was ignored")
			}
			assertReport(t, stderr.Bytes(), "refused", "create.input")
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Error("vetoed save published a plan")
			}
		} else {
			if exit != 0 {
				t.Error("complete save paired with EOF was lost")
			}
			assertReport(t, stderr.Bytes(), "written", "create.written")
			body, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
				t.Error("complete save at EOF did not preserve reviewed bytes")
			}
		}
		preserved(t, root)
	}
}

type noProgressReader struct{ calls int }

func (r *noProgressReader) Read([]byte) (int, error) {
	r.calls++
	if r.calls > 1 {
		return 0, errors.New("reader retried after no progress")
	}
	return 0, nil
}

func TestZeroProgressIsRefusalRatherThanRetryOrEOF(t *testing.T) {
	root, target := sandbox(t)
	input := &noProgressReader{}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr) != 1 || input.calls != 1 {
		t.Error("zero progress was retried or treated as successful EOF")
	}
	assertReport(t, stderr.Bytes(), "refused", "create.input")
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Error("zero progress created a plan")
	}
	preserved(t, root)
}

func TestLiteralActionWordsAreNotGlobalCommandsInsideValuePrompts(t *testing.T) {
	for _, word := range []string{"save", "cancel", "remove", "review", "back"} {
		root, target := sandbox(t)
		var out, stderr bytes.Buffer
		input := "add\nrepository-access\nrepository-a\nbind\nrepository-a\nrepository_scope\ntext\n" + word + "\nreview\nsave\n"
		if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
			t.Fatalf("literal field word %q was treated as an action", word)
		}
		plan := savedPlan(t, target)
		answers, ok := plan["answers"].(map[string]any)
		if !ok {
			t.Fatal("literal field answer absent")
		}
		sections, ok := answers["sections"].(map[string]any)
		if !ok {
			t.Fatal("literal section answer absent")
		}
		local, ok := sections["repository-a"].(map[string]any)
		if !ok || local["repository_scope"] != word {
			t.Errorf("literal field changed: %#v", local)
		}
		preserved(t, root)
	}
}
