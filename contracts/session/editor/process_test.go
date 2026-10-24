//go:build darwin || linux

package editor_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("OBK_EDITOR_CONTRACT_NATIVE_LOCK") == "1" {
		file, err := os.Open(os.Getenv("OBK_EDITOR_CONTRACT_NATIVE_ROOT"))
		if err != nil {
			os.Exit(91)
		}
		var held syscall.Stat_t
		if syscall.Fstat(int(file.Fd()), &held) != nil || syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
			os.Exit(92)
		}
		fmt.Printf("LOCKED %d %d %d\n", os.Getpid(), uint64(held.Dev), uint64(held.Ino))
		var one [1]byte
		_, _ = os.Stdin.Read(one[:])
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestActualSecondProcessNativeExclusionRefusesBeforeIdentity(t *testing.T) {
	s := session(t, planBody, recordBody)
	before := snapshot(t, s.base)
	native := nativeObjectAt(t, s.root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^$")
	command.Env = append(os.Environ(), "OBK_EDITOR_CONTRACT_NATIVE_LOCK=1", "OBK_EDITOR_CONTRACT_NATIVE_ROOT="+s.root)
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		_ = input.Close()
		if !waited {
			cancel()
			_ = command.Wait()
		}
	}()
	line, err := bufio.NewReader(output).ReadString('\n')
	if err != nil || len(line) > 256 {
		t.Fatal("native child proof missing")
	}
	fields := strings.Fields(line)
	if len(fields) != 4 || fields[0] != "LOCKED" || fields[1] != strconv.Itoa(command.Process.Pid) || fields[2] != strconv.FormatUint(native.Device, 10) || fields[3] != strconv.FormatUint(native.Inode, 10) {
		t.Fatal("same-owned-PID/native-object lock proof mismatched")
	}
	probe, err := os.Open(s.root)
	if err != nil {
		t.Fatal(err)
	}
	busy := syscall.Flock(int(probe.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	_ = probe.Close()
	if busy != syscall.EWOULDBLOCK && busy != syscall.EAGAIN {
		t.Fatal("actual second-process lock not independently proved busy")
	}
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, "cancel\n")
	report(t, got, "edit.context", 1, false)
	if clock.calls != 0 || got.consumed != 0 {
		t.Fatal("native exclusion loser consumed identity/time")
	}
	unchanged(t, s, before)
	_, _ = input.Write([]byte("x"))
	_ = input.Close()
	if err := command.Wait(); err != nil {
		t.Fatal(err)
	}
	waited = true
}
