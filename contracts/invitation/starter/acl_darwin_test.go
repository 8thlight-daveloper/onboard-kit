package starter_test

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func setExplicitEmptyACL(t *testing.T, path string) {
	t.Helper()
	present, _, _, _ := nativeACL(t, path)
	if present {
		t.Fatal("empty ACL fixture must begin with actual ACL absence")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	var file *os.File
	if info.IsDir() {
		file, err = os.Open(path)
	} else {
		file, err = os.OpenFile(path, os.O_RDWR, 0)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if file.Close() != nil {
			t.Error("native ACL writer close failed")
		}
	}()
	attrs := nativeAttributes{BitmapCount: 5, Common: 0x00400000}
	payload := make([]byte, 52)
	binary.LittleEndian.PutUint32(payload[:4], 8)
	binary.LittleEndian.PutUint32(payload[4:8], 44)
	binary.LittleEndian.PutUint32(payload[8:12], 0x012cc16d)
	_, _, failure := syscall.Syscall6(syscall.SYS_FSETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&payload[0])), uintptr(len(payload)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attrs)
	runtime.KeepAlive(payload)
	if failure == syscall.ENOTSUP || failure == syscall.ENOSYS {
		t.Skipf("native empty ACL fixture unsupported: errno %d", failure)
	}
	if failure != 0 {
		t.Fatalf("native empty ACL fixture failed: errno %d", failure)
	}
	present, encoded, _, _ := nativeACL(t, path)
	body, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || !present || len(body) != 44 || binary.LittleEndian.Uint32(body[36:40]) != 0 {
		t.Fatal("host did not establish explicit zero-entry ACL distinct from absence")
	}
}

func ownedChmod(t *testing.T, args ...string) {
	t.Helper()
	out, err := exec.Command("/bin/chmod", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("owned native chmod fixture failed; output bytes=%d", len(out))
	}
}

func TestDarwinExistingACLNoopAndReplacement(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64/amd64")
	}
	for _, kind := range []string{"explicit-empty", "deny-execute"} {
		for _, changed := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "/retain", true: "/replace"}[changed], func(t *testing.T) {
				r := makeRecovery(t, true)
				if kind == "explicit-empty" {
					setExplicitEmptyACL(t, r.record())
				} else {
					ownedChmod(t, "+a", "everyone deny execute", r.record())
				}
				held := objectGuard(t, r.record())
				if !held.Security.AccessPresent {
					t.Fatal("native ACL presence not reached")
				}
				r.value.Roles[3].BeforeGuard = &held
				if changed {
					r.value.Roles[3].Action = "reconcile"
					after := append(fixture(t, "begun.yaml"), []byte("# ACL-preserving reviewed bytes.\n")...)
					r.value.Roles[3].AfterBase64 = encoded(after)
					r.value.Roles[3].AfterSHA256 = digest(after)
				}
				rewritePending(t, &r)
				clock := 0
				got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
				after := objectGuard(t, r.record())
				if after.Security != held.Security {
					t.Error("existing ACL presence/payload or UID/GID/mode was weakened")
				}
				if !changed {
					wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
					if after != held {
						t.Error("ACL no-op required replacement capability or replaced inode")
					}
					assertIdle(t, r.sandbox)
				} else if got.report.Status == "opened" {
					if got.exit != 0 {
						t.Errorf("completed ACL replacement needs repair: %s", got.report.Code)
					}
					body, err := base64.StdEncoding.DecodeString(r.value.Roles[3].AfterBase64)
					if err != nil {
						t.Fatal(err)
					}
					assertBytes(t, r.record(), body)
					if after.Inode == held.Inode && after.Device == held.Device {
						t.Error("ACL replacement did not replace pathname inode")
					}
					assertIdle(t, r.sandbox)
				} else {
					if got.exit != 1 || got.report.Status != "pending" {
						t.Error("ACL preserve-or-refuse lost pending truth")
					}
					if after != held {
						t.Error("refused ACL replacement changed original inode/controls")
					}
					assertBytes(t, r.record(), fixture(t, "begun.yaml"))
				}
				if clock != 0 {
					t.Error("native ACL recovery sampled event")
				}
			})
		}
	}
}

func hasInheritedReadGrant(t *testing.T, encoded string, inherited bool) bool {
	t.Helper()
	body, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(body) < 44 {
		t.Fatal("invalid observed native ACL payload")
	}
	count := binary.LittleEndian.Uint32(body[36:40])
	if count == ^uint32(0) {
		return false
	}
	for i := uint32(0); i < count; i++ {
		start := uint64(44) + uint64(i)*24
		if start+24 > uint64(len(body)) {
			t.Fatal("invalid observed native ACE bounds")
		}
		flags := binary.LittleEndian.Uint32(body[start+16 : start+20])
		rights := binary.LittleEndian.Uint32(body[start+20 : start+24])
		if flags&0xf == 1 && rights&(1<<1) != 0 {
			if inherited && flags&(1<<4) != 0 {
				return true
			}
			if !inherited && flags&(1<<5) != 0 {
				return true
			}
		}
	}
	return false
}

func TestDarwinInheritedGrantCannotBroadenPrivateStagingOrPublication(t *testing.T) {
	r := makeRecovery(t, false)
	parent := filepath.Dir(r.record())
	ownedChmod(t, "+a", "everyone allow read,file_inherit,directory_inherit", parent)
	parentGuard := objectGuard(t, parent)
	if !parentGuard.Security.AccessPresent || !hasInheritedReadGrant(t, parentGuard.Security.AccessBase64, false) {
		t.Fatal("native inheritable parent grant not reached")
	}
	control := filepath.Join(parent, "owned-inheritance-control")
	put(t, control, []byte("disposable control\n"))
	controlGuard := objectGuard(t, control)
	if controlGuard.Security.Mode != "0600" || !controlGuard.Security.AccessPresent || !hasInheritedReadGrant(t, controlGuard.Security.AccessBase64, true) {
		t.Fatal("native creation did not prove inherited read despite mode0600")
	}
	if err := os.Remove(control); err != nil {
		t.Fatal(err)
	}
	for _, i := range []int{1, 3} {
		r.value.Roles[i].Parents[1].Guard = &parentGuard
	}
	rewritePending(t, &r)
	before := snapshot(t, r.base)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
	if objectGuard(t, parent) != parentGuard {
		t.Error("publication stripped or changed existing parent ACL")
	}
	if got.exit == 0 {
		wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
		assertIdle(t, r.sandbox)
	} else if got.exit != 1 || got.report.Status != "pending" {
		t.Error("inherited grant refusal lost pending truth")
	}
	actual := snapshot(t, r.base)
	for rel, item := range actual {
		if _, existed := before[rel]; existed || item.Mode&os.ModeSymlink != 0 {
			continue
		}
		holdsBytes := item.Mode.IsRegular() && item.Bytes != ""
		if item.Mode.IsDir() {
			for child, value := range actual {
				if strings.HasPrefix(child, rel+string(os.PathSeparator)) && value.Mode.IsRegular() && value.Bytes != "" {
					holdsBytes = true
				}
			}
		}
		if got.exit != 0 && !holdsBytes {
			continue
		}
		path := filepath.Join(r.base, rel)
		g := objectGuard(t, path)
		if g.Security.AccessPresent || g.Security.DefaultPresent {
			t.Error("new private object holding bytes inherited access grants")
		}
		if item.Mode.IsRegular() && g.Security.Mode != "0600" {
			t.Error("new private file does not have mode0600")
		}
		if item.Mode.IsDir() && g.Security.Mode != "0700" {
			t.Error("new private directory does not have mode0700")
		}
	}
	if clock != 0 {
		t.Error("inherited grant recovery sampled event")
	}
}

func TestDarwinIdleExplicitEmptyACLIsNotSafeAbsence(t *testing.T) {
	s := setup(t)
	if err := os.MkdirAll(s.pending(), 0700); err != nil {
		t.Fatal(err)
	}
	setExplicitEmptyACL(t, s.pending())
	before := snapshot(t, s.base)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\n")
	if got.exit != 1 || got.report.Code != "start.context" || clock != 0 || strings.Contains(got.out, "Review complete.") {
		t.Error("ACL-bearing fixed idle namespace was accepted or altered")
	}
	unchanged(t, s, before)
}
