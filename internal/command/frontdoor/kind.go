package frontdoor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/tui"
)

const (
	planByteLimit     = 16 << 20
	archiveSignature  = "PK\x03\x04"
	wrongKindCode     = "format.wrong-kind"
	notYAMLCode       = "file.not-yaml"
	notAPlanCode      = "file.not-a-plan"
	expectedSomething = "but a plan or an invitation is expected"
)

var namedKinds = map[string]string{"preparation": "a preparation", "record": "a record"}

func Check(path string) (tui.FileCheck, []model.Problem) {
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return refused("%s does not exist.", path), nil
	case err != nil:
		return refused("%s cannot be read: %v.", path, err), nil
	case info.IsDir():
		return refused("%s is a folder, "+expectedSomething+".", path), nil
	case !info.Mode().IsRegular():
		return refused("%s is not a regular file, "+expectedSomething+".", path), nil
	}
	data, err := readBounded(path)
	if err != nil {
		return refused("%s cannot be read: %v.", path, err), nil
	}
	full := absolute(path)
	if bytes.HasPrefix(data, []byte(archiveSignature)) {
		return tui.FileCheck{Kind: tui.FileInvitation, Path: full}, nil
	}
	if len(data) > planByteLimit {
		return refused("%s is larger than 16 MiB, "+expectedSomething+".", path), nil
	}
	_, problems := decoder.New().Decode(data)
	first, failed := firstError(problems)
	switch {
	case !failed:
		return tui.FileCheck{Kind: tui.FilePlan, Path: full}, nil
	case first.Code == wrongKindCode && namedKinds[first.Detail["found"]] != "":
		return refused("%s is "+namedKinds[first.Detail["found"]]+", "+expectedSomething+".", path), nil
	case first.Code == wrongKindCode, first.Code == notYAMLCode, first.Code == notAPlanCode:
		return refused("%s is neither a plan nor an invitation.", path), nil
	}
	return tui.FileCheck{Kind: tui.FileRefused, Path: full, Refusal: fmt.Sprintf("%s is a plan that cannot be opened: %s.", path, first.Message)}, problems
}

func refused(format string, path string, rest ...any) tui.FileCheck {
	return tui.FileCheck{Kind: tui.FileRefused, Path: path, Refusal: fmt.Sprintf(format, append([]any{path}, rest...)...)}
}

func firstError(problems []model.Problem) (model.Problem, bool) {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return problem, true
		}
	}
	return model.Problem{}, false
}

func readBounded(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, planByteLimit+1))
}

func absolute(path string) string {
	if full, err := filepath.Abs(path); err == nil {
		return full
	}
	return path
}
