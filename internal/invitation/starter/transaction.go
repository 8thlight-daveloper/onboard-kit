package starter

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

type transaction struct {
	namespace    *directory
	owner        ownerIntent
	pending      *pending
	pendingBytes []byte
	claim        publicationClaim
	receipt      *receipt
	states       map[string]snapshot
	root, home   *chain
}

func (t *transaction) close() error {
	if t.namespace != nil {
		err := t.namespace.close()
		t.namespace = nil
		return err
	}
	return nil
}
func atomicProtocol(d *directory, name string, data []byte, previous snapshot) (snapshot, error) {
	current, err := readAt(d.fd(), name, len(data)+33554432, true)
	if err != nil || !sourceEqual(current, previous) {
		return snapshot{}, errContext
	}
	stage := "." + strings.TrimSuffix(name, ".json") + ".next"
	file, g, err := createFile(d, stage)
	if err != nil {
		return snapshot{}, err
	}
	if err := writeAll(file, string(data)); err != nil {
		file.Close()
		return snapshot{}, err
	}
	if err := file.Close(); err != nil {
		return snapshot{}, errContext
	}
	staged, err := readAt(d.fd(), stage, len(data), true)
	if err != nil || !staged.present || !reflect.DeepEqual(staged.guard, &g) || !bytes.Equal(staged.data, data) {
		return snapshot{}, errContext
	}
	if previous.present {
		err = unix.Renameat(d.fd(), stage, d.fd(), name)
	} else {
		err = publishAbsent(d.fd(), stage, d.fd(), name)
	}
	if err != nil {
		return snapshot{}, errContext
	}
	return readAt(d.fd(), name, len(data), true)
}
func (t *transaction) write(name string, value any, maximum int) error {
	data := encode(value)
	if len(data) > maximum {
		return errLimit
	}
	next, err := atomicProtocol(t.namespace, name, data, t.states[name])
	if err != nil {
		return err
	}
	t.states[name] = next
	return nil
}
func inspectTransaction(root, home *chain) (*transaction, error) {
	if len(home.missing) > 0 {
		return nil, nil
	}
	d, err := dirAt(home.last().fd(), "import-pending", filepath.Join(home.absolute, "import-pending"))
	if errors.Is(err, unix.ENOENT) {
		return nil, nil
	}
	if err != nil {
		return nil, errContext
	}
	fail := func() (*transaction, error) { d.close(); return nil, errContext }
	if !private(d.original) {
		return fail()
	}
	names, err := entries(d, 8)
	if err != nil {
		d.close()
		return nil, err
	}
	if len(names) == 0 {
		d.close()
		return nil, nil
	}
	t := &transaction{namespace: d, states: map[string]snapshot{}, root: root, home: home}
	for _, entry := range names {
		name := entry.Name()
		maximum := 16384
		switch name {
		case "pending.json":
			maximum = 33554432
		case "owner.json", "publication.json", "complete.json", ".pending.next":
			if name == ".pending.next" {
				maximum = 33554432
			}
		default:
			return fail()
		}
		s, err := readAt(d.fd(), name, maximum, true)
		if err != nil || !s.present {
			return fail()
		}
		t.states[name] = s
	}
	if s := t.states["complete.json"]; s.present {
		var r receipt
		if decode(s.data, &r, 16384) != nil || !validReceipt(r) || r.RootDevice != root.last().original.Device || r.RootInode != root.last().original.Inode || r.HomeDevice != home.last().original.Device || r.HomeInode != home.last().original.Inode || r.NamespaceDevice != d.original.Device || r.NamespaceInode != d.original.Inode || r.NamespaceSecuritySHA256 != securityHash(d.original.Security) {
			return fail()
		}
		t.receipt = &r
	}
	if s := t.states["owner.json"]; s.present {
		if decode(s.data, &t.owner, 16384) != nil || !validOwner(t.owner) || !reflect.DeepEqual(t.owner.Namespace, d.original) || t.owner.RootDevice != root.last().original.Device || t.owner.RootInode != root.last().original.Inode {
			return fail()
		}
	} else if t.receipt == nil {
		return fail()
	}
	if s := t.states["pending.json"]; s.present {
		var p pending
		if err := decode(s.data, &p, 33554432); err != nil {
			return fail()
		}
		if err := p.validate(); err != nil {
			d.close()
			return nil, err
		}
		if p.Transaction != t.owner.Transaction && t.states["owner.json"].present || !sameObject(p.RootGuard, root.last().original) || !sameObject(p.HomeGuard, home.last().original) || p.Home != home.absolute {
			return fail()
		}
		if t.states["owner.json"].present && t.owner.PendingSHA256 != hash(s.data) {
			return fail()
		}
		if t.receipt != nil && (t.receipt.Transaction != p.Transaction || t.receipt.PendingSHA256 != hash(s.data) || t.receipt.SessionKey != p.SessionKey) {
			return fail()
		}
		t.pending = &p
		t.pendingBytes = s.data
		expected := targetRelative(p.SessionKey, p.Record, p.Preparation)
		for i, r := range p.Roles {
			if len(r.Parents) != len(strings.Split(expected[i], "/"))-1 {
				return fail()
			}
		}
		if t.receipt == nil && (!reflect.DeepEqual(p.RootGuard, root.last().original) || !reflect.DeepEqual(p.HomeGuard, home.last().original)) {
			return fail()
		}
	}
	if s := t.states["publication.json"]; s.present {
		if decode(s.data, &t.claim, 16384) != nil || !validClaim(t.claim) || t.pending != nil && (t.claim.Transaction != t.pending.Transaction || t.claim.PendingSHA256 != hash(t.pendingBytes)) || t.pending == nil && t.receipt == nil {
			return fail()
		}
		if t.receipt != nil && (t.claim.Transaction != t.receipt.Transaction || t.claim.PendingSHA256 != t.receipt.PendingSHA256) {
			return fail()
		}
	}
	if t.pending != nil && !t.states["publication.json"].present {
		t.claim = emptyClaim(t.pending.Transaction, hash(t.pendingBytes))
	}
	if s := t.states[".pending.next"]; s.present {
		if t.owner.PendingTemporary == nil || !reflect.DeepEqual(s.guard, t.owner.PendingTemporary) {
			return fail()
		}
	}
	if t.pending == nil && t.receipt == nil && t.owner.PendingTemporary != nil && !t.states[".pending.next"].present {
		return fail()
	}
	return t, nil
}
func validOwner(o ownerIntent) bool {
	return hex64(o.Transaction) && decimal(o.RootDevice) && decimal(o.RootInode) && validGuard(o.Namespace, "directory") && private(o.Namespace) && (o.PendingSHA256 == "" || hex64(o.PendingSHA256)) && (o.PendingTemporary == nil || validGuard(*o.PendingTemporary, "regular") && private(*o.PendingTemporary))
}
func validClaim(c publicationClaim) bool {
	if !hex64(c.Transaction) || !hex64(c.PendingSHA256) || len(c.Roles) != 4 {
		return false
	}
	for i, r := range c.Roles {
		if r.Role != roleNames()[i] || r.Staging != nil && (!validGuard(*r.Staging, "directory") || !private(*r.Staging)) || r.Temporary != nil && (!validGuard(*r.Temporary, "regular") || !private(*r.Temporary) || r.Staging == nil) {
			return false
		}
		last := -1
		for _, p := range r.NewParents {
			n, err := strconv.Atoi(p.Ordinal)
			if err != nil || !decimal(p.Ordinal) || n <= last || !validGuard(p.Guard, "directory") || !private(p.Guard) {
				return false
			}
			last = n
		}
	}
	return true
}
func validReceipt(r receipt) bool {
	if !hex64(r.Transaction) || !hex64(r.SessionKey) || !hex64(r.PendingSHA256) || !decimal(r.RootDevice) || !decimal(r.RootInode) || !decimal(r.HomeDevice) || !decimal(r.HomeInode) || !decimal(r.NamespaceDevice) || !decimal(r.NamespaceInode) || !hex64(r.NamespaceSecuritySHA256) || len(r.Roles) != 4 {
		return false
	}
	for i, v := range r.Roles {
		if v.Role != roleNames()[i] || !hex64(v.SHA256) || !decimal(v.Device) || !decimal(v.Inode) || !hex64(v.SecuritySHA256) {
			return false
		}
	}
	return true
}
func emptyClaim(token, digest string) publicationClaim {
	c := publicationClaim{Version: 1, Transaction: token, PendingSHA256: digest, Roles: []roleClaim{}}
	for _, name := range roleNames() {
		c.Roles = append(c.Roles, roleClaim{Role: name, NewParents: []createdParent{}})
	}
	return c
}
func startTransaction(root, home *chain, p pending) (*transaction, error) {
	data := encode(p)
	if len(data) > 33554432 {
		return nil, errLimit
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	d, err := dirAt(home.last().fd(), "import-pending", filepath.Join(home.absolute, "import-pending"))
	if errors.Is(err, unix.ENOENT) {
		d, err = createDirectory(home.last(), "import-pending")
	}
	if err != nil {
		return nil, err
	}
	if !private(d.original) {
		d.close()
		return nil, errContext
	}
	values, err := entries(d, 0)
	if err != nil || len(values) != 0 {
		d.close()
		return nil, errContext
	}
	t := &transaction{namespace: d, states: map[string]snapshot{}, root: root, home: home, pending: &p, pendingBytes: data, claim: emptyClaim(p.Transaction, hash(data))}
	t.owner = ownerIntent{Version: 1, Transaction: p.Transaction, RootDevice: p.RootGuard.Device, RootInode: p.RootGuard.Inode, Namespace: d.original}
	if err := t.write("owner.json", t.owner, 16384); err != nil {
		return t, err
	}
	file, g, err := createFile(d, ".pending.next")
	if err != nil {
		return t, err
	}
	t.owner.PendingTemporary = &g
	t.owner.PendingSHA256 = hash(data)
	if err := t.write("owner.json", t.owner, 16384); err != nil {
		file.Close()
		return t, err
	}
	if err := writeAll(file, string(data)); err != nil {
		file.Close()
		return t, err
	}
	if err := file.Close(); err != nil {
		return t, errContext
	}
	s, err := readAt(d.fd(), ".pending.next", 33554432, true)
	if err != nil || !s.present || !reflect.DeepEqual(s.guard, &g) || !bytes.Equal(s.data, data) {
		return t, errContext
	}
	if err := publishAbsent(d.fd(), ".pending.next", d.fd(), "pending.json"); err != nil {
		return t, errContext
	}
	t.states["pending.json"] = s
	if err := t.write("publication.json", t.claim, 16384); err != nil {
		return t, err
	}
	return t, nil
}
func (t *transaction) roleBase(i int) *directory {
	if i == 0 || i == 2 {
		return t.home.last()
	}
	return t.root.last()
}
func (t *transaction) rolePath(i int) string {
	p := t.pending
	return targetRelative(p.SessionKey, p.Record, p.Preparation)[i]
}
func (t *transaction) claimedParent(path string) (*objectGuard, error) {
	var found *objectGuard
	for i, role := range t.claim.Roles {
		parts := strings.Split(t.rolePath(i), "/")
		for _, p := range role.NewParents {
			n, err := strconv.Atoi(p.Ordinal)
			if err != nil || n >= len(parts)-1 || t.pending.Roles[i].Parents[n].Present {
				return nil, errContext
			}
			absolute := filepath.Join(t.roleBase(i).path, filepath.Join(parts[:n+1]...))
			if absolute == path {
				if found != nil && !reflect.DeepEqual(*found, p.Guard) {
					return nil, errContext
				}
				g := p.Guard
				found = &g
			}
		}
	}
	return found, nil
}
func (t *transaction) parent(i int, create bool) (*directory, []*directory, error) {
	current := t.roleBase(i)
	var owned []*directory
	parts := strings.Split(t.rolePath(i), "/")
	fail := func() (*directory, []*directory, error) {
		for _, d := range owned {
			d.close()
		}
		return nil, nil, errContext
	}
	for j, name := range parts[:len(parts)-1] {
		expected := t.pending.Roles[i].Parents[j]
		d, err := dirAt(current.fd(), name, filepath.Join(current.path, name))
		if expected.Present {
			if err != nil {
				return fail()
			}
			owned = append(owned, d)
			if !reflect.DeepEqual(d.original, *expected.Guard) {
				return fail()
			}
		} else {
			claim, claimErr := t.claimedParent(filepath.Join(current.path, name))
			if claimErr != nil {
				if d != nil {
					d.close()
				}
				return fail()
			}
			if err == nil {
				owned = append(owned, d)
				if claim == nil || !reflect.DeepEqual(d.original, *claim) {
					return fail()
				}
			} else if errors.Is(err, unix.ENOENT) && create && claim == nil {
				d, err = createDirectory(current, name)
				if err != nil {
					return fail()
				}
				owned = append(owned, d)
				t.claim.Roles[i].NewParents = append(t.claim.Roles[i].NewParents, createdParent{Ordinal: strconv.Itoa(j), Guard: d.original})
				if err := t.write("publication.json", t.claim, 16384); err != nil {
					return fail()
				}
			} else {
				return fail()
			}
		}
		current = d
	}
	return current, owned, nil
}
func closeDirectories(values []*directory) error {
	var result error
	for i := len(values) - 1; i >= 0; i-- {
		if err := values[i].close(); err != nil {
			result = err
		}
	}
	return result
}
func (t *transaction) targetState(i int, parent *directory) (snapshot, string, error) {
	role := t.pending.Roles[i]
	s, err := readAt(parent.fd(), filepath.Base(t.rolePath(i)), 2097152, false)
	if err != nil {
		return s, "", err
	}
	before, _ := base64Bytes(role.BeforeBase64, 2097152)
	after, _ := base64Bytes(role.AfterBase64, 2097152)
	if s.present && role.BeforePresent && reflect.DeepEqual(s.guard, role.BeforeGuard) && bytes.Equal(s.data, before) {
		if role.Action == "retain" {
			return s, "done", nil
		}
		return s, "before", nil
	}
	if !s.present && !role.BeforePresent {
		return s, "before", nil
	}
	claim := t.claim.Roles[i]
	if s.present && claim.Temporary != nil && sameObject(*s.guard, *claim.Temporary) && bytes.Equal(s.data, after) {
		security := claim.Temporary.Security
		if role.BeforePresent {
			security = role.BeforeGuard.Security
		}
		if reflect.DeepEqual(s.guard.Security, security) {
			return s, "done", nil
		}
	}
	return s, "", errContext
}
func (t *transaction) verifyContext() error {
	if err := t.root.recheck(); err != nil {
		return err
	}
	if err := t.home.recheck(); err != nil {
		return err
	}
	g, err := guard(t.namespace.fd())
	if err != nil || !reflect.DeepEqual(g, t.owner.Namespace) {
		return errContext
	}
	for name, s := range t.states {
		current, err := readAt(t.namespace.fd(), name, 33554432, true)
		if err != nil || !sourceEqual(current, s) {
			return errContext
		}
	}
	return nil
}
func (t *transaction) publishRole(i int) error {
	if err := t.verifyContext(); err != nil {
		return err
	}
	parent, owned, err := t.parent(i, true)
	if err != nil {
		return err
	}
	defer closeDirectories(owned)
	_, state, err := t.targetState(i, parent)
	if err != nil || state == "done" {
		return err
	}
	r := t.pending.Roles[i]
	c := &t.claim.Roles[i]
	name := ".obk-start-" + t.pending.Transaction + "-" + r.Role
	var staging *directory
	if c.Staging == nil {
		staging, err = createDirectory(parent, name)
		if err != nil {
			return err
		}
		g := staging.original
		c.Staging = &g
		if err := t.write("publication.json", t.claim, 16384); err != nil {
			staging.close()
			return err
		}
	} else {
		staging, err = dirAt(parent.fd(), name, filepath.Join(parent.path, name))
		if err != nil || !reflect.DeepEqual(staging.original, *c.Staging) {
			if staging != nil {
				staging.close()
			}
			return errContext
		}
	}
	defer staging.close()
	after, _ := base64Bytes(r.AfterBase64, 2097152)
	final := security{}
	if r.BeforePresent {
		final = r.BeforeGuard.Security
		if final.AccessPresent || final.DefaultPresent {
			return errContext
		}
	}
	var file *os.File
	if c.Temporary == nil {
		f, g, e := createFile(staging, "payload")
		if e != nil {
			return e
		}
		file = f
		c.Temporary = &g
		if err := t.write("publication.json", t.claim, 16384); err != nil {
			file.Close()
			return err
		}
		if !r.BeforePresent {
			final = g.Security
		}
		if err := writeAll(file, string(after)); err != nil {
			file.Close()
			return err
		}
		if err := applySecurity(int(file.Fd()), final); err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return errContext
		}
	} else {
		s, err := readAt(staging.fd(), "payload", 2097152, false)
		if err != nil || !s.present || !sameObject(*s.guard, *c.Temporary) {
			return errContext
		}
		if !r.BeforePresent {
			final = c.Temporary.Security
		}
		if !reflect.DeepEqual(s.guard.Security, c.Temporary.Security) && !reflect.DeepEqual(s.guard.Security, final) {
			return errContext
		}
		if !bytes.Equal(s.data, after) || !reflect.DeepEqual(s.guard.Security, final) {
			fd, err := unix.Openat(staging.fd(), "payload", unix.O_WRONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			if err != nil {
				return errContext
			}
			file = os.NewFile(uintptr(fd), "payload")
			g, err := guard(fd)
			if err != nil || !sameObject(g, *c.Temporary) {
				file.Close()
				return errContext
			}
			if err := unix.Ftruncate(fd, 0); err != nil {
				file.Close()
				return errContext
			}
			if err := writeAll(file, string(after)); err != nil {
				file.Close()
				return err
			}
			if err := applySecurity(fd, final); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return errContext
			}
		}
	}
	if err := t.verifyContext(); err != nil {
		return err
	}
	_, state, err = t.targetState(i, parent)
	if err != nil || state != "before" {
		return errContext
	}
	staged, err := readAt(staging.fd(), "payload", 2097152, false)
	if err != nil || !staged.present || !sameObject(*staged.guard, *c.Temporary) || !bytes.Equal(staged.data, after) || !reflect.DeepEqual(staged.guard.Security, final) {
		return errContext
	}
	target := filepath.Base(t.rolePath(i))
	if r.BeforePresent {
		err = unix.Renameat(staging.fd(), "payload", parent.fd(), target)
	} else {
		err = publishAbsent(staging.fd(), "payload", parent.fd(), target)
	}
	if err != nil {
		return errContext
	}
	return nil
}
func (t *transaction) rollForward() error {
	if t.pending == nil || t.receipt != nil {
		return errContext
	}
	if err := t.preflightRoles(); err != nil {
		return err
	}
	if !t.states["publication.json"].present {
		if err := t.write("publication.json", t.claim, 16384); err != nil {
			return err
		}
	}
	for i := range t.pending.Roles {
		if err := t.publishRole(i); err != nil {
			return err
		}
	}
	if err := t.verifyContext(); err != nil {
		return err
	}
	p := t.pending
	r := receipt{Version: 1, Transaction: p.Transaction, SessionKey: p.SessionKey, PendingSHA256: hash(t.pendingBytes), RootDevice: p.RootGuard.Device, RootInode: p.RootGuard.Inode, HomeDevice: p.HomeGuard.Device, HomeInode: p.HomeGuard.Inode, NamespaceDevice: t.namespace.original.Device, NamespaceInode: t.namespace.original.Inode, NamespaceSecuritySHA256: securityHash(t.namespace.original.Security), Roles: []receiptRole{}}
	for i := range p.Roles {
		parent, owned, err := t.parent(i, false)
		if err != nil {
			return err
		}
		s, state, err := t.targetState(i, parent)
		closeDirectories(owned)
		if err != nil || state != "done" {
			return errContext
		}
		r.Roles = append(r.Roles, receiptRole{Role: p.Roles[i].Role, SHA256: hash(s.data), Device: s.guard.Device, Inode: s.guard.Inode, SecuritySHA256: securityHash(s.guard.Security)})
	}
	if err := t.write("complete.json", r, 16384); err != nil {
		return err
	}
	t.receipt = &r
	return nil
}

func (t *transaction) preflightRoles() error {
	for i, role := range t.pending.Roles {
		if role.Action != "retain" && role.BeforePresent && (role.BeforeGuard.Security.AccessPresent || role.BeforeGuard.Security.DefaultPresent) {
			return errContext
		}
		current := t.roleBase(i)
		parts := strings.Split(t.rolePath(i), "/")
		var opened []*directory
		missing := false
		for j, name := range parts[:len(parts)-1] {
			expected := role.Parents[j]
			d, err := dirAt(current.fd(), name, filepath.Join(current.path, name))
			if errors.Is(err, unix.ENOENT) && !expected.Present {
				claim, claimErr := t.claimedParent(filepath.Join(current.path, name))
				if claimErr != nil || claim != nil {
					closeDirectories(opened)
					return errContext
				}
				missing = true
				break
			}
			if err != nil {
				closeDirectories(opened)
				return errContext
			}
			opened = append(opened, d)
			if expected.Present {
				if !reflect.DeepEqual(d.original, *expected.Guard) {
					closeDirectories(opened)
					return errContext
				}
			} else {
				claim, err := t.claimedParent(d.path)
				if err != nil || claim == nil || !reflect.DeepEqual(d.original, *claim) {
					closeDirectories(opened)
					return errContext
				}
			}
			current = d
		}
		claim := t.claim.Roles[i]
		if missing {
			closeDirectories(opened)
			if role.BeforePresent || claim.Staging != nil || claim.Temporary != nil {
				return errContext
			}
			continue
		}
		_, state, err := t.targetState(i, current)
		if err != nil {
			closeDirectories(opened)
			return err
		}
		name := ".obk-start-" + t.pending.Transaction + "-" + role.Role
		d, err := dirAt(current.fd(), name, filepath.Join(current.path, name))
		if errors.Is(err, unix.ENOENT) {
			closeDirectories(opened)
			if claim.Staging != nil {
				return errContext
			}
			continue
		}
		if err != nil || claim.Staging == nil || !reflect.DeepEqual(d.original, *claim.Staging) {
			if d != nil {
				d.close()
			}
			closeDirectories(opened)
			return errContext
		}
		items, err := entries(d, 1)
		if err != nil {
			d.close()
			closeDirectories(opened)
			return err
		}
		if len(items) == 0 {
			if claim.Temporary != nil && state != "done" {
				d.close()
				closeDirectories(opened)
				return errContext
			}
		} else {
			if items[0].Name() != "payload" || claim.Temporary == nil {
				d.close()
				closeDirectories(opened)
				return errContext
			}
			s, err := readAt(d.fd(), "payload", 2097152, false)
			if err != nil || !s.present || !sameObject(*s.guard, *claim.Temporary) {
				d.close()
				closeDirectories(opened)
				return errContext
			}
			final := claim.Temporary.Security
			if role.BeforePresent {
				final = role.BeforeGuard.Security
			}
			if !reflect.DeepEqual(s.guard.Security, claim.Temporary.Security) && !reflect.DeepEqual(s.guard.Security, final) {
				d.close()
				closeDirectories(opened)
				return errContext
			}
		}
		d.close()
		if err := closeDirectories(opened); err != nil {
			return err
		}
	}
	return nil
}
func removeGuarded(parent *directory, name string, expected objectGuard, directory bool) error {
	flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC | unix.O_NONBLOCK
	if directory {
		flags |= unix.O_DIRECTORY
	}
	fd, err := unix.Openat(parent.fd(), name, flags, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	if err != nil {
		return errContext
	}
	actual, err := guard(fd)
	unix.Close(fd)
	if err != nil || !reflect.DeepEqual(actual, expected) {
		return errContext
	}
	unlinkFlags := 0
	if directory {
		unlinkFlags = unix.AT_REMOVEDIR
	}
	if err := unix.Unlinkat(parent.fd(), name, unlinkFlags); err != nil {
		return errContext
	}
	return nil
}
func (t *transaction) cleanup() error {
	if t.receipt == nil && t.pending != nil {
		return errContext
	}
	if t.pending != nil && t.states["publication.json"].present {
		for i, c := range t.claim.Roles {
			if c.Staging == nil {
				continue
			}
			parent, owned, err := t.parent(i, false)
			if err != nil {
				return err
			}
			name := ".obk-start-" + t.pending.Transaction + "-" + c.Role
			d, err := dirAt(parent.fd(), name, filepath.Join(parent.path, name))
			if errors.Is(err, unix.ENOENT) {
				closeDirectories(owned)
				continue
			}
			if err != nil || !reflect.DeepEqual(d.original, *c.Staging) {
				if d != nil {
					d.close()
				}
				closeDirectories(owned)
				return errContext
			}
			values, err := entries(d, 1)
			if err != nil {
				d.close()
				closeDirectories(owned)
				return err
			}
			if len(values) > 0 {
				if values[0].Name() != "payload" || c.Temporary == nil {
					d.close()
					closeDirectories(owned)
					return errContext
				}
				s, err := readAt(d.fd(), "payload", 2097152, false)
				if err != nil || !s.present || !sameObject(*s.guard, *c.Temporary) {
					d.close()
					closeDirectories(owned)
					return errContext
				}
				security := c.Temporary.Security
				if t.pending.Roles[i].BeforePresent {
					security = t.pending.Roles[i].BeforeGuard.Security
				}
				if !reflect.DeepEqual(s.guard.Security, c.Temporary.Security) && !reflect.DeepEqual(s.guard.Security, security) {
					d.close()
					closeDirectories(owned)
					return errContext
				}
				if err := removeGuarded(d, "payload", *s.guard, false); err != nil {
					d.close()
					closeDirectories(owned)
					return err
				}
			}
			d.close()
			if err := removeGuarded(parent, name, *c.Staging, true); err != nil {
				closeDirectories(owned)
				return err
			}
			if err := closeDirectories(owned); err != nil {
				return err
			}
		}
	}
	for _, name := range []string{"pending.json", ".pending.next", "publication.json", "owner.json", "complete.json"} {
		s := t.states[name]
		if !s.present {
			continue
		}
		if name == ".pending.next" && (t.owner.PendingTemporary == nil || !reflect.DeepEqual(*s.guard, *t.owner.PendingTemporary)) {
			return errContext
		}
		current, err := readAt(t.namespace.fd(), name, 33554432, true)
		if err != nil || !sourceEqual(current, s) {
			return errContext
		}
		if err := removeGuarded(t.namespace, name, *s.guard, false); err != nil {
			return err
		}
		delete(t.states, name)
	}
	values, err := entries(t.namespace, 0)
	if err != nil || len(values) > 0 {
		return errContext
	}
	return nil
}
