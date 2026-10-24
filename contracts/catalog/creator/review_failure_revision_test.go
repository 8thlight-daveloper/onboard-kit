package creator_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type reviewFailure struct {
	storage bytes.Buffer
	mode    int
	failed  bool
}

func (w *reviewFailure) Write(p []byte) (int, error) {
	if w.failed {
		return 0, errors.New("owned display failure")
	}
	joined := append(append([]byte(nil), w.storage.Bytes()...), p...)
	if bytes.Contains(joined, []byte("Review complete.")) {
		w.failed = true
		if w.mode == 0 {
			return 0, errors.New("owned display failure")
		}
		if w.mode == 1 {
			accepted := len(p) - 1
			if accepted < 0 {
				accepted = 0
			}
			w.storage.Write(p[:accepted])
			return accepted, nil
		}
		w.storage.Write(p)
		return len(p), errors.New("owned display failure")
	}
	return w.storage.Write(p)
}

func TestReviewFailureFixtureReachesAllFaultsThroughWriteAndWriteString(t *testing.T) {
	for _, method := range []string{"Write", "io.WriteString"} {
		for mode := 0; mode < 3; mode++ {
			writer := &reviewFailure{mode: mode}
			write := func(text string) (int, error) {
				if method == "io.WriteString" {
					return io.WriteString(writer, text)
				}
				return writer.Write([]byte(text))
			}
			prefix := "accepted display prefix\n"
			if n, err := write(prefix); n != len(prefix) || err != nil || writer.failed {
				t.Fatalf("fixture rejected safe prefix method=%s mode=%d n=%d err=%v", method, mode, n, err)
			}
			part := "Review com"
			if n, err := write(part); n != len(part) || err != nil || writer.failed {
				t.Fatalf("fixture rejected pre-boundary fragment method=%s mode=%d n=%d err=%v", method, mode, n, err)
			}
			boundary := "plete. Choose save, back, or cancel:\n"
			n, err := write(boundary)
			wantCount := len(boundary)
			wantError := true
			if mode == 0 {
				wantCount = 0
			}
			if mode == 1 {
				wantCount = len(boundary) - 1
				wantError = false
			}
			if !writer.failed || n != wantCount || (err != nil) != wantError {
				t.Errorf("fixture fault not reached method=%s mode=%d n=%d err=%v failed=%v", method, mode, n, err, writer.failed)
			}
			wantStored := prefix + part + boundary[:wantCount]
			if writer.storage.String() != wantStored {
				t.Errorf("accepted display bytes differ method=%s mode=%d got=%q want=%q", method, mode, writer.storage.String(), wantStored)
			}
			if n, err := write("must remain rejected"); n != 0 || err == nil || writer.storage.String() != wantStored {
				t.Errorf("fixture did not remain failed method=%s mode=%d n=%d err=%v", method, mode, n, err)
			}
		}
	}
}
