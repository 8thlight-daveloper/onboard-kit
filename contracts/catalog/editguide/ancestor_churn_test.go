package editguide_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestSaveSurvivesEntriesComingAndGoingInAnAncestorDirectory(t *testing.T) {
	const saves = 60
	first, _ := setup(t, 0700)
	shared := filepath.Dir(first)
	stop := make(chan struct{})
	var churn sync.WaitGroup
	churn.Add(1)
	go func() {
		defer churn.Done()
		for n := 0; ; n++ {
			select {
			case <-stop:
				return
			default:
			}
			name := filepath.Join(shared, "churn-"+strconv.Itoa(n))
			if os.Mkdir(name, 0700) == nil {
				os.Remove(name)
			}
		}
	}()
	defer func() {
		close(stop)
		churn.Wait()
	}()
	for i := 0; i < saves; i++ {
		_, target := setup(t, 0700)
		input := &countedInput{source: fixture(t, "retire.txt")}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 0 {
			t.Fatalf("save %d refused while an ancestor directory gained and lost entries: %q", i, stderr.String())
		}
		saved, err := os.ReadFile(target)
		if err != nil || !bytes.Equal(saved, fixture(t, "retired.yaml")) {
			t.Fatalf("save %d did not write the reviewed bytes: %q", i, saved)
		}
	}
}
