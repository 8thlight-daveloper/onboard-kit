package creator_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type raceReader struct {
	source   []byte
	position int
	entered  bool
	ready    chan<- struct{}
	release  <-chan struct{}
	abort    <-chan struct{}
}

func (r *raceReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	select {
	case <-r.abort:
		return 0, io.EOF
	default:
	}
	if r.position == len("review\n") && !r.entered {
		r.entered = true
		r.ready <- struct{}{}
		<-r.release
	}
	select {
	case <-r.abort:
		return 0, io.EOF
	default:
	}
	if r.position == len(r.source) {
		return 0, io.EOF
	}
	p[0] = r.source[r.position]
	r.position++
	return 1, nil
}

func TestConcurrentReviewedCreationPublishesExactlyOneCompletePlan(t *testing.T) {
	logicalRoot, err := os.MkdirTemp("", "catalog-creator-race-")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.EvalSymlinks(logicalRoot)
	if err != nil {
		if cleanupErr := os.RemoveAll(logicalRoot); cleanupErr != nil {
			t.Errorf("owned sandbox resolution cleanup failed: %v", cleanupErr)
		}
		t.Fatal(err)
	}
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	abort := make(chan struct{})
	finished := make(chan struct{}, 2)
	allDone := make(chan struct{})
	var releaseOnce, abortOnce sync.Once
	var done sync.WaitGroup
	releaseWorkers := func() { releaseOnce.Do(func() { close(release) }) }
	abortWorkers := func() { abortOnce.Do(func() { close(abort) }); releaseWorkers() }
	workersStarted := false
	t.Cleanup(func() {
		abortWorkers()
		if workersStarted {
			timer := time.NewTimer(10 * time.Second)
			defer timer.Stop()
			select {
			case <-allDone:
			case <-timer.C:
				t.Errorf("workers did not finish after bounded release; exclusively owned sandbox retained at %s", root)
				return
			}
		}
		if err := os.RemoveAll(root); err != nil {
			t.Errorf("owned race sandbox cleanup failed at %s: %v", root, err)
		}
	})
	if err := os.Chown(root, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"preparation.obk.yaml", "record.obk.yaml", "organization.obk.yaml"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("PRESERVE ORIGINAL\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(root, "created.obk.yaml")
	type result struct {
		exit   int
		stderr bytes.Buffer
	}
	results := make([]result, 2)
	done.Add(2)
	workersStarted = true
	for i := range 2 {
		go func(i int) {
			defer done.Done()
			input := &raceReader{source: []byte("review\nsave\n"), ready: ready, release: release, abort: abort}
			results[i].exit = newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, io.Discard, &results[i].stderr)
			finished <- struct{}{}
		}(i)
	}
	go func() { done.Wait(); close(allDone) }()
	barrierTimer := time.NewTimer(10 * time.Second)
	defer barrierTimer.Stop()
	for entered := 0; entered < 2; {
		select {
		case <-ready:
			entered++
		case <-finished:
			abortWorkers()
			t.Fatal("both creators must reach a complete review before the publication race")
		case <-barrierTimer.C:
			abortWorkers()
			t.Fatal("creators did not reach the bounded review barrier")
		}
	}
	releaseWorkers()
	completionTimer := time.NewTimer(10 * time.Second)
	defer completionTimer.Stop()
	select {
	case <-allDone:
	case <-completionTimer.C:
		abortWorkers()
		t.Fatal("reviewed publication workers did not complete within the bound")
	}
	written := 0
	for _, result := range results {
		if result.exit == 0 {
			written++
			assertReport(t, result.stderr.Bytes(), "written", "create.written")
		} else {
			if result.exit != 1 {
				t.Errorf("unexpected competing exit: %d", result.exit)
			}
			assertReport(t, result.stderr.Bytes(), "refused", "create.stale")
		}
	}
	if written != 1 {
		t.Errorf("wanted one exclusive publication, got %d", written)
	}
	body, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
		t.Error("exclusive publication did not retain one complete reviewed plan")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Errorf("competing temporary remained: %v", entries)
	}
	preserved(t, root)
}
