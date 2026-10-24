package editguide_test

import (
	"io"
	"testing"
)

func TestDisplayFaultFixtureControlsBothWriteEntryPaths(t *testing.T) {
	for _, stringPath := range []bool{false, true} {
		for mode := 0; mode < 3; mode++ {
			writer := &displayFailure{mode: mode}
			write := func(value string) (int, error) {
				if stringPath {
					return io.WriteString(writer, value)
				}
				return writer.Write([]byte(value))
			}
			if n, err := write("Review com"); n != 10 || err != nil {
				t.Fatal("fixture prefix did not succeed")
			}
			boundary := "plete. Choose save, back, or cancel:\n"
			n, err := write(boundary)
			want := len(boundary)
			wantError := true
			if mode == 0 {
				want = 0
			}
			if mode == 1 {
				want--
				wantError = false
			}
			if !writer.failed || n != want || (err != nil) != wantError || writer.data.String() != "Review com"+boundary[:want] {
				t.Errorf("display fault control not reached string=%v mode=%d n=%d err=%v", stringPath, mode, n, err)
			}
		}
	}
}
