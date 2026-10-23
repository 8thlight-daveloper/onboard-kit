package catalog_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestSingleHyphenIDValuesRemainLiteralUnknownIDs(t *testing.T) {
	t.Parallel()
	var cases []argumentCase
	readFixture(t, "single-hyphen-arguments.json", &cases)
	if len(cases) != 4 {
		t.Fatal("expected four independently specified single-hyphen argument cases")
	}
	shared := newCommand()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Source == "" {
				t.Fatal("missing argument fixture provenance")
			}
			before := make([]string, len(c.Args))
			copy(before, c.Args)
			got := run(shared, c.Args)
			if got.exit != c.Exit || got.out != "" {
				t.Fatalf("single-hyphen ID is a literal unknown ID, not missing flag data: %#v", got)
			}
			assertReport(t, got.err, c.Code)
			if !reflect.DeepEqual(c.Args, before) {
				t.Fatal("literal ID refusal mutated caller arguments")
			}
		})
	}
}

func TestExactArgumentBoundariesRejectUsage(t *testing.T) {
	t.Parallel()
	var cases []argumentCase
	readFixture(t, "exact-argument-boundaries.json", &cases)
	if len(cases) != 15 {
		t.Fatal("expected fifteen independently specified exact argument boundary cases")
	}
	shared := newCommand()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Source == "" {
				t.Fatal("missing boundary fixture provenance")
			}
			before := make([]string, len(c.Args))
			copy(before, c.Args)
			got := run(shared, c.Args)
			if got.exit != c.Exit || got.out != "" {
				t.Fatalf("exact syntax boundary must refuse usage without result/help bytes: %#v", got)
			}
			assertReport(t, got.err, c.Code)
			if !reflect.DeepEqual(c.Args, before) {
				t.Fatal("usage refusal mutated caller arguments")
			}
		})
	}
}

type fullCountErrorWriter struct {
	content bytes.Buffer
	writes  int
}

func (writer *fullCountErrorWriter) Write(data []byte) (int, error) {
	writer.writes++
	count, _ := writer.content.Write(data)
	return count, errors.New("PRIVATE FULL COUNT WRITER FAILURE")
}

func (writer *fullCountErrorWriter) Len() int {
	return writer.content.Len()
}

func (writer *fullCountErrorWriter) String() string {
	return writer.content.String()
}

func TestFullCountWriterErrorRefusesResultAndHelpSuccess(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		args []string
		code string
	}{
		{name: "list", args: []string{"list"}, code: "catalog.result-output"},
		{name: "show", args: []string{"show", "--id", "company-email"}, code: "catalog.result-output"},
		{name: "export", args: []string{"export", "--id", "pair-assign"}, code: "catalog.result-output"},
		{name: "help", args: []string{"--help"}, code: "catalog.help-output"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout fullCountErrorWriter
			var stderr bytes.Buffer
			exit := newCommand().Run(c.args, &stdout, &stderr)
			if stdout.Len() == 0 {
				t.Fatal("full-count error control never received result bytes")
			}
			if exit != 1 {
				t.Fatalf("a full byte count does not cancel a writer error: got exit%d", exit)
			}
			assertReport(t, stderr.String(), c.code)
		})
	}
}

func TestFullCountRefusalReportErrorReturnsOne(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		args []string
		code string
	}{
		{name: "usage", args: []string{}, code: "catalog.usage"},
		{name: "unknown", args: []string{"show", "--id", "unknown-private-selector"}, code: "catalog.unknown-id"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr fullCountErrorWriter
			exit := newCommand().Run(c.args, &stdout, &stderr)
			if exit != 1 || stdout.Len() != 0 {
				t.Fatalf("failed refusal report must return1 without stdout, got exit%d stdout%d", exit, stdout.Len())
			}
			assertReport(t, stderr.String(), c.code)
		})
	}
}

func TestUnusedFullCountErrorStderrDoesNotFailSuccess(t *testing.T) {
	t.Parallel()
	var stderr fullCountErrorWriter
	if newCommand().Run([]string{"list"}, io.Discard, &stderr) != 0 || stderr.writes != 0 {
		t.Fatal("unused stderr must not fail a successful stdout-only operation")
	}
}
