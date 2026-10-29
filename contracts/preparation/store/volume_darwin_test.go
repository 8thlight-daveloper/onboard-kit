package store_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

func attachVolume(t *testing.T, filesystem string) string {
	t.Helper()
	return attachVolumeAt(t, filesystem, "")
}

func attachVolumeAt(t *testing.T, filesystem, mount string) string {
	t.Helper()
	if testing.Short() {
		t.Skip("disk image volumes are slow, and a short run leaves them out")
	}
	hdiutil, err := exec.LookPath("hdiutil")
	if err != nil {
		t.Skip("hdiutil is not available, so no disk image volume can be made")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(base, "volume.dmg")
	if value, err := exec.Command(hdiutil, "create", "-size", "8m", "-fs", filesystem, "-volname", "OBKTEST", "-layout", "NONE", image).CombinedOutput(); err != nil {
		t.Skipf("hdiutil cannot create a %s image: %v: %s", filesystem, err, value)
	}
	if mount == "" {
		mount = filepath.Join(base, "mount")
	}
	if value, err := exec.Command(hdiutil, "attach", "-nobrowse", "-noverify", "-mountpoint", mount, image).CombinedOutput(); err != nil {
		t.Skipf("hdiutil cannot attach a %s image: %v: %s", filesystem, err, value)
	}
	t.Cleanup(func() {
		for attempt := 0; attempt < attempts; attempt++ {
			if err := exec.Command(hdiutil, "detach", "-force", mount).Run(); err == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Errorf("cannot detach %s", mount)
	})
	return mount
}

func volumeSandbox(t *testing.T, mount, before string) sandbox {
	t.Helper()
	root := filepath.Join(mount, "root")
	result := sandbox{Base: mount, Root: root, Plan: filepath.Join(root, "organization.yaml"), Prep: filepath.Join(root, filepath.FromSlash(destination))}
	writeFile(t, result.Plan, fixture(t, "plan.yaml"))
	if before != "" {
		writeFile(t, result.Prep, fixture(t, before))
	}
	return result
}

func TestDarwinVolumeThatCannotKeepAFilePrivateRefusesSaveWithItsOwnCode(t *testing.T) {
	for _, filesystem := range []string{"MS-DOS", "ExFAT"} {
		for _, c := range []struct {
			name, before, reviewed string
			directory, consent     bool
			want                   contract.Code
		}{
			{name: "replace", before: "identity.yaml", reviewed: "named.yaml", want: contract.CodeVolume},
			{name: "new-in-existing-directory", reviewed: "named.yaml", directory: true, want: contract.CodeVolume},
			{name: "new-directories-with-consent", reviewed: "identity.yaml", consent: true, want: contract.CodeVolume},
			{name: "unchanged-draft-needs-no-write", before: "identity.yaml", reviewed: "identity.yaml", want: contract.CodeUnchanged},
		} {
			t.Run(filesystem+"/"+c.name, func(t *testing.T) {
				mount := attachVolume(t, filesystem)
				box := volumeSandbox(t, mount, c.before)
				if c.directory {
					if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
						t.Fatal(err)
					}
				}
				listed := newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan})
				wantCode(t, listed.Code, contract.CodeListed)
				if want := c.before != ""; (len(listed.People) == 1) != want {
					t.Fatalf("list on the volume gave %#v", listed.People)
				}
				if len(listed.People) == 1 && (listed.People[0].Account != account || listed.People[0].Destination != destination || listed.People[0].Problem != "") {
					t.Fatalf("list on the volume gave %#v", listed.People)
				}
				loaded := box.load(t)
				if loaded.Exists() != (c.before != "") {
					t.Fatalf("load on the volume reports Exists %v", loaded.Exists())
				}
				held := snapshot(t, box.Root)
				result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, c.reviewed), CreateDirectories: c.consent})
				wantCode(t, result.Code, c.want)
				if result.Committed || result.CreatedDirectories {
					t.Fatalf("committed %v created directories %v", result.Committed, result.CreatedDirectories)
				}
				if after := snapshot(t, box.Root); !sameMap(after, held) {
					t.Fatalf("refused save changed the volume: before %v after %v", held, after)
				}
			})
		}
	}
}

func TestDarwinVolumeMountedWhereTheFileIsWrittenRefusesSaveWithItsOwnCode(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "")
	person := filepath.Dir(box.Prep)
	if err := os.MkdirAll(person, 0o700); err != nil {
		t.Fatal(err)
	}
	attachVolumeAt(t, "MS-DOS", person)
	loaded := box.load(t)
	if loaded.Exists() || loaded.DirectoriesMissing() {
		t.Fatalf("load reports Exists %v DirectoriesMissing %v", loaded.Exists(), loaded.DirectoriesMissing())
	}
	held := snapshot(t, box.Root)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
	wantCode(t, result.Code, contract.CodeVolume)
	if result.Committed || result.CreatedDirectories {
		t.Fatalf("committed %v created directories %v", result.Committed, result.CreatedDirectories)
	}
	if after := snapshot(t, box.Root); !sameMap(after, held) {
		t.Fatalf("refused save changed the volume: before %v after %v", held, after)
	}
}
