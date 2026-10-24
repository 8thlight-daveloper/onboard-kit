package starter

import (
	"bytes"
	"encoding/base64"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/sys/unix"
	reviewapi "onboard-kit/contracts/invitation/reviewer"
)

type registered struct {
	metadata  metadata
	snapshot  snapshot
	directory objectGuard
}
type registry struct {
	entries []registered
	present bool
	guard   objectGuard
}

func scanRegistry(home *chain) (registry, error) {
	result := registry{entries: []registered{}}
	if len(home.missing) > 0 {
		return result, nil
	}
	d, err := dirAt(home.last().fd(), "sessions", filepath.Join(home.absolute, "sessions"))
	if errors.Is(err, unix.ENOENT) {
		return result, nil
	}
	if err != nil {
		return result, errContext
	}
	defer d.close()
	if !private(d.original) {
		return result, errContext
	}
	result.present = true
	result.guard = d.original
	entries, err := entries(d, 128)
	if err != nil {
		return result, errContext
	}
	total := 0
	for _, entry := range entries {
		if !hex64(entry.Name()) {
			return result, errContext
		}
		folder, err := dirAt(d.fd(), entry.Name(), filepath.Join(d.path, entry.Name()))
		if err != nil {
			return result, errContext
		}
		if !private(folder.original) {
			folder.close()
			return result, errContext
		}
		names, err := entriesInSession(folder)
		if err != nil || !names {
			folder.close()
			return result, errContext
		}
		body, err := readAt(folder.fd(), "session.json", 2097152, true)
		folderGuard := folder.original
		folder.close()
		if err != nil {
			return result, err
		}
		if !body.present {
			return result, errContext
		}
		total += len(body.data)
		if total > 16777216 {
			return result, errContext
		}
		var m metadata
		if err := decode(body.data, &m, 2097152); err != nil {
			return result, err
		}
		if !m.valid() || m.SessionKey != entry.Name() {
			return result, errContext
		}
		result.entries = append(result.entries, registered{metadata: m, snapshot: body, directory: folderGuard})
	}
	return result, nil
}
func entriesInSession(d *directory) (bool, error) {
	values, err := entries(d, 2)
	if err != nil {
		return false, err
	}
	return len(values) == 2 && values[0].Name() == "plan.obk.yaml" && values[1].Name() == "session.json", nil
}

type contextSelection struct {
	key      string
	root     string
	matching *metadata
	others   []metadata
}

func chooseContext(incoming reviewapi.Incoming, root *chain, home *chain, r registry) (contextSelection, error) {
	chosen := contextSelection{key: sessionKey(incoming.PlanID, incoming.Files.Account, root.absolute), root: root.absolute, others: []metadata{}}
	for _, entry := range r.entries {
		m := entry.metadata
		plan, err := base64Bytes(m.PlanIDBase64, 1048576)
		if err != nil {
			return chosen, errContext
		}
		exact := string(plan) == incoming.PlanID && m.Account == incoming.Files.Account
		old, err := openChain(m.Root)
		if err != nil || len(old.missing) > 0 {
			if old != nil {
				old.close()
			}
			if exact {
				return chosen, errContext
			}
			continue
		}
		g := old.last().original
		old.close()
		if g.Device != m.RootDevice || g.Inode != m.RootInode {
			if exact {
				return chosen, errContext
			}
			continue
		}
		same := sameObject(g, root.last().original)
		if same && exact {
			if chosen.matching != nil || m.Record != incoming.Paths.Record || m.Preparation != incoming.Paths.Preparation {
				return chosen, errContext
			}
			copy := m
			chosen.matching = &copy
			chosen.key = m.SessionKey
			chosen.root = m.Root
		} else {
			for _, a := range []string{filepath.Join(root.absolute, incoming.Paths.Record), filepath.Join(root.absolute, incoming.Paths.Preparation)} {
				for _, b := range []string{filepath.Join(m.Root, m.Record), filepath.Join(m.Root, m.Preparation)} {
					if pathConflict(a, b) {
						return chosen, errContext
					}
					shared, err := sharedDestinationParent(a, b)
					if err != nil || shared {
						return chosen, errContext
					}
					sa, ea := absoluteObject(a)
					sb, eb := absoluteObject(b)
					if ea != nil || eb != nil {
						return chosen, errContext
					}
					if sa != nil && sb != nil && sameObject(*sa, *sb) {
						return chosen, errContext
					}
				}
			}
			if exact {
				chosen.others = append(chosen.others, m)
			}
		}
	}
	sort.Slice(chosen.others, func(i, j int) bool { return chosen.others[i].SessionKey < chosen.others[j].SessionKey })
	return chosen, nil
}
func pathConflict(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
func sharedDestinationParent(a, b string) (bool, error) {
	if filepath.Base(a) != filepath.Base(b) {
		return false, nil
	}
	first, err := openChain(filepath.Dir(a))
	if err != nil {
		return false, err
	}
	defer first.close()
	second, err := openChain(filepath.Dir(b))
	if err != nil {
		return false, err
	}
	defer second.close()
	return len(first.missing) == 0 && len(second.missing) == 0 && sameObject(first.last().original, second.last().original), nil
}
func absoluteObject(path string) (*objectGuard, error) {
	c, err := openChain(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer c.close()
	if len(c.missing) > 0 {
		return nil, nil
	}
	fd, err := unix.Openat(c.last().fd(), filepath.Base(path), unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil, nil
	}
	if err != nil {
		return nil, errContext
	}
	defer unix.Close(fd)
	g, err := guard(fd)
	if err != nil || g.Kind != "regular" {
		return nil, errContext
	}
	return &g, nil
}

func sameSelection(a, b contextSelection) bool { return reflect.DeepEqual(a, b) }
func targetPaths(home, root, key, record, preparation string) []string {
	return []string{filepath.Join(home, "sessions", key, "plan.obk.yaml"), filepath.Join(root, preparation), filepath.Join(home, "sessions", key, "session.json"), filepath.Join(root, record)}
}
func targetRelative(key, record, preparation string) []string {
	return []string{filepath.Join("sessions", key, "plan.obk.yaml"), preparation, filepath.Join("sessions", key, "session.json"), record}
}
func snapshots(root, home *chain, key, record, preparation string) ([]snapshot, error) {
	paths := targetRelative(key, record, preparation)
	result := make([]snapshot, 4)
	for i, path := range paths {
		base := root.last()
		if i == 0 || i == 2 {
			if len(home.missing) > 0 {
				parts := strings.Split(path, "/")
				p := make([]parentGuard, len(parts)-1)
				for j := range p {
					p[j] = parentGuard{Ordinal: strconvOrdinal(j)}
				}
				result[i] = snapshot{parents: p}
				continue
			}
			base = home.last()
		}
		maximum := 1048576
		if i == 2 {
			maximum = 2097152
		}
		s, err := readTarget(base, path, maximum, i == 0 || i == 2)
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	for i, a := range result {
		if !a.present {
			continue
		}
		for j := 0; j < i; j++ {
			b := result[j]
			if b.present && sameObject(*a.guard, *b.guard) {
				return nil, errContext
			}
		}
	}
	return result, nil
}
func strconvOrdinal(n int) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}
func namespaceChecks(root, home string, key, record, preparation, archive string, ss []snapshot, source snapshot, r registry) error {
	paths := targetPaths(home, root, key, record, preparation)
	for i, a := range paths {
		if len(a) > 4096 {
			return errLimit
		}
		for j := 0; j < i; j++ {
			if pathConflict(a, paths[j]) {
				return errContext
			}
		}
		if archive != "" && pathConflict(a, archive) {
			return errContext
		}
		if pathConflict(a, filepath.Join(home, "import-pending")) {
			return errContext
		}
		if (i == 1 || i == 3) && pathConflict(a, filepath.Join(home, "sessions")) {
			return errContext
		}
		if ss[i].present && source.present && sameObject(*ss[i].guard, *source.guard) {
			return errContext
		}
		for _, entry := range r.entries {
			if ss[i].present && entry.snapshot.present && i != 2 && sameObject(*ss[i].guard, *entry.snapshot.guard) {
				return errContext
			}
		}
	}
	if archive != "" && (pathConflict(archive, filepath.Join(home, "sessions")) || pathConflict(archive, filepath.Join(home, "import-pending"))) {
		return errContext
	}
	return nil
}
func sameRegistry(a, b registry) bool { return reflect.DeepEqual(a, b) }
func metadataProposal(incoming reviewapi.Incoming, selection contextSelection, root objectGuard, token string, previous snapshot) ([]byte, string, error) {
	m := metadata{Version: 1, SessionKey: selection.key, PlanIDBase64: base64.StdEncoding.EncodeToString([]byte(incoming.PlanID)), Account: incoming.Files.Account, Root: selection.root, RootDevice: root.Device, RootInode: root.Inode, Record: incoming.Paths.Record, Preparation: incoming.Paths.Preparation, PlanSHA256: incoming.Reference.Hash, Transaction: token}
	action := "install"
	if previous.present {
		var old metadata
		if decode(previous.data, &old, 2097152) != nil {
			return nil, "", errContext
		}
		old.Transaction = token
		if reflect.DeepEqual(old, m) {
			return previous.data, "retain", nil
		}
		action = "replace"
	}
	data := encode(m)
	if len(data) > 2097152 {
		return nil, "", errLimit
	}
	return data, action, nil
}
func sameSources(a, b []snapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !snapshotEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}
func sourceEqual(a, b snapshot) bool {
	return a.present == b.present && bytes.Equal(a.data, b.data) && reflect.DeepEqual(a.guard, b.guard)
}
