package editor

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	reviewcontract "onboard-kit/contracts/session/reviewer"
)

type nativeGuard struct {
	Device, Inode  uint64
	UID, GID, Mode uint32
	Attributes     nativeAttributes
}
type held struct {
	file  *os.File
	guard nativeGuard
	path  string
}
type snapshot struct {
	held
	parent *os.File
	name   string
	body   []byte
}
type registration struct {
	key, planID, account, root, record, preparation string
	device, inode                                   uint64
}
type context struct {
	args                      arguments
	selected                  registration
	root, home                *os.File
	files                     []*os.File
	guards                    map[string]nativeGuard
	metadata                  map[string][]byte
	plan, preparation, record snapshot
	locked                    []*os.File
	groups                    map[uint32]bool
	source                    reviewcontract.Source
	absences                  []absence
}

type absence struct {
	ParentPath        string
	Parent            nativeGuard
	Remaining         []string
	ComparisonProfile int
}

type registryEntry struct {
	registration   registration
	metadata, plan snapshot
}

func (c *context) close() bool {
	good := c.unlock()
	for i := len(c.files) - 1; i >= 0; i-- {
		if c.files[i].Close() != nil {
			good = false
		}
	}
	c.files = nil
	return good
}
func (c *context) unlock() bool {
	good := true
	for i := len(c.locked) - 1; i >= 0; i-- {
		if unix.Flock(int(c.locked[i].Fd()), unix.LOCK_UN) != nil {
			good = false
		}
	}
	c.locked = nil
	return good
}
func (c *context) lock() bool {
	files := []*os.File{c.root, c.home}
	sort.Slice(files, func(i, j int) bool {
		a, b := c.guards[files[i].Name()], c.guards[files[j].Name()]
		if a.Device != b.Device {
			return a.Device < b.Device
		}
		return a.Inode < b.Inode
	})
	var previous nativeGuard
	for i, f := range files {
		guard := c.guards[f.Name()]
		if i > 0 && guard.Device == previous.Device && guard.Inode == previous.Inode {
			continue
		}
		if unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB) != nil {
			c.unlock()
			return false
		}
		c.locked = append(c.locked, f)
		previous = guard
	}
	return true
}
func guard(file *os.File) (nativeGuard, bool) {
	var stat unix.Stat_t
	if unix.Fstat(int(file.Fd()), &stat) != nil {
		return nativeGuard{}, false
	}
	attrs, ok := readNativeAttributes(int(file.Fd()))
	if !ok {
		return nativeGuard{}, false
	}
	return nativeGuard{Device: uint64(stat.Dev), Inode: uint64(stat.Ino), UID: stat.Uid, GID: stat.Gid, Mode: uint32(stat.Mode) & 07777, Attributes: attrs}, true
}
func (c *context) private(g nativeGuard, mode uint32) bool {
	return g.UID == uint32(os.Geteuid()) && c.groups[g.GID] && g.Mode == mode && !g.Attributes.AccessPresent && !g.Attributes.DefaultPresent
}
func (c *context) own(file *os.File, path string) (nativeGuard, bool) {
	g, ok := guard(file)
	if !ok {
		_ = file.Close()
		return nativeGuard{}, false
	}
	c.files = append(c.files, file)
	c.guards[path] = g
	return g, true
}
func (c *context) directory(base *os.File, name, path string) (*os.File, bool) {
	fd, err := unix.Openat(int(base.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, false
	}
	file := os.NewFile(uintptr(fd), path)
	_, ok := c.own(file, path)
	return file, ok
}
func (c *context) absolute(path string) (*os.File, bool) {
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, false
	}
	current := os.NewFile(uintptr(fd), "/")
	if _, ok := c.own(current, "/"); !ok {
		return nil, false
	}
	built := "/"
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}
		built = filepath.Join(built, part)
		next, ok := c.directory(current, part, built)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}
func (c *context) participant(base *os.File, relative string) (snapshot, bool) {
	if !relativePath(relative) {
		return snapshot{}, false
	}
	parts := strings.Split(relative, "/")
	parent := base
	path := base.Name()
	for _, part := range parts[:len(parts)-1] {
		path = filepath.Join(path, part)
		next, ok := c.directory(parent, part, path)
		if !ok {
			return snapshot{}, false
		}
		parent = next
	}
	name := parts[len(parts)-1]
	path = filepath.Join(path, name)
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return snapshot{}, false
	}
	f := os.NewFile(uintptr(fd), path)
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG {
		_ = f.Close()
		return snapshot{}, false
	}
	g, ok := c.own(f, path)
	if !ok {
		return snapshot{}, false
	}
	return snapshot{held: held{file: f, guard: g, path: path}, parent: parent, name: name}, true
}
func boundedBody(file *os.File, limit int) ([]byte, bool) {
	data, code := readBoundedBody(file, limit)
	return data, code == ""
}
func readBoundedBody(file *os.File, limit int) ([]byte, string) {
	var before unix.Stat_t
	if unix.Fstat(int(file.Fd()), &before) != nil || before.Mode&unix.S_IFMT != unix.S_IFREG || before.Size < 0 {
		return nil, "source"
	}
	if before.Size > int64(limit) {
		return nil, "limit"
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, "source"
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if len(data) > limit {
		return nil, "limit"
	}
	if err != nil {
		return nil, "source"
	}
	return data, ""
}
func relativePath(path string) bool {
	if path == "" || len(path) > 4096 || !utf8.ValidString(path) || strings.ContainsAny(path, "\x00\\") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
func entries(file *os.File, maximum int) ([]string, bool) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, false
	}
	names, err := file.Readdirnames(maximum + 1)
	if err != nil && err != io.EOF || len(names) > maximum {
		return nil, false
	}
	return names, true
}
func openContext(args arguments) (*context, string) {
	c := &context{args: args, guards: map[string]nativeGuard{}, metadata: map[string][]byte{}, groups: map[uint32]bool{uint32(os.Getegid()): true}}
	fail := func(code string) (*context, string) { c.close(); return nil, code }
	groups, err := os.Getgroups()
	if err != nil {
		return fail("context")
	}
	for _, group := range groups {
		if group >= 0 {
			c.groups[uint32(group)] = true
		}
	}
	var ok bool
	c.root, ok = c.absolute(args.root)
	if !ok {
		return fail("context")
	}
	c.home, ok = c.absolute(args.home)
	if !ok || !c.private(c.guards[c.home.Name()], 0700) {
		return fail("context")
	}
	if !c.lock() {
		return fail("context")
	}
	if !c.pending() {
		return fail("pending")
	}
	sessions, ok := c.directory(c.home, "sessions", filepath.Join(c.home.Name(), "sessions"))
	if !ok || !c.private(c.guards[sessions.Name()], 0700) {
		return fail("context")
	}
	keys, ok := entries(sessions, 128)
	if !ok {
		return fail("context")
	}
	sort.Strings(keys)
	total := 0
	selected := false
	participants := []snapshot{}
	registry := []registryEntry{}
	for _, key := range keys {
		if !hexKey(key) {
			return fail("context")
		}
		session, ok := c.directory(sessions, key, filepath.Join(sessions.Name(), key))
		if !ok || !c.private(c.guards[session.Name()], 0700) {
			return fail("context")
		}
		names, ok := entries(session, 2)
		if !ok || len(names) != 2 {
			return fail("context")
		}
		sort.Strings(names)
		if names[0] != "plan.obk.yaml" || names[1] != "session.json" {
			return fail("context")
		}
		metadata, ok := c.participant(session, "session.json")
		if !ok || !c.private(metadata.guard, 0600) {
			return fail("context")
		}
		data, ok := boundedBody(metadata.file, 2097152)
		if !ok {
			return fail("context")
		}
		total += len(data)
		if total > 16777216 {
			return fail("context")
		}
		after, ok := guard(metadata.file)
		if !ok || !reflect.DeepEqual(metadata.guard, after) {
			return fail("context")
		}
		c.metadata[metadata.path] = data
		registration, ok := parseRegistration(data, key)
		if !ok {
			return fail("context")
		}
		plan, ok := c.participant(session, "plan.obk.yaml")
		if !ok || !c.private(plan.guard, 0600) {
			return fail("context")
		}
		registry = append(registry, registryEntry{registration: registration, metadata: metadata, plan: plan})
		if key == args.key {
			c.selected = registration
			selected = true
		}
	}
	if !selected {
		return fail("context")
	}
	for _, entry := range registry {
		registration := entry.registration
		matching := registration.planID == c.selected.planID && registration.account == c.selected.account
		registeredRoot, missing, ok := c.registeredRoot(registration.root)
		if !ok || missing != nil && matching {
			return fail("context")
		}
		row := []snapshot{entry.metadata, entry.plan}
		var prep, record snapshot
		var rootGuard nativeGuard
		if missing != nil {
			for _, relative := range []string{registration.preparation, registration.record} {
				proof := *missing
				proof.Remaining = append(append([]string(nil), missing.Remaining...), strings.Split(relative, "/")...)
				c.absences = append(c.absences, proof)
			}
		} else {
			rootGuard = c.guards[registeredRoot.Name()]
			if rootGuard.Device != registration.device || rootGuard.Inode != registration.inode {
				return fail("context")
			}
			for i, relative := range []string{registration.preparation, registration.record} {
				participant, missing, ok := c.optionalParticipant(registeredRoot, relative)
				if !ok || missing != nil && matching {
					return fail("context")
				}
				if missing != nil {
					c.absences = append(c.absences, *missing)
					continue
				}
				row = append(row, participant)
				if i == 0 {
					prep = participant
				} else {
					record = participant
				}
			}
		}
		for _, participant := range row {
			for _, other := range participants {
				if participant.guard.Device == other.guard.Device && participant.guard.Inode == other.guard.Inode || strings.HasPrefix(participant.path, other.path+string(filepath.Separator)) || strings.HasPrefix(other.path, participant.path+string(filepath.Separator)) {
					return fail("context")
				}
			}
			participants = append(participants, participant)
		}
		if registration.key == args.key {
			selectedRoot := c.guards[c.root.Name()]
			if selectedRoot.Device != rootGuard.Device || selectedRoot.Inode != rootGuard.Inode {
				return fail("context")
			}
			c.selected = registration
			c.plan = entry.plan
			c.preparation = prep
			c.record = record
		}
	}
	if !selected {
		return fail("context")
	}
	for _, p := range participants {
		for path, g := range c.guards {
			if path == p.path {
				continue
			}
			if p.guard.Device == g.Device && p.guard.Inode == g.Inode {
				return fail("context")
			}
		}
	}
	if !c.disjointAbsences(participants) || !c.absencesFresh() {
		return fail("context")
	}
	if !c.unlock() {
		return fail("context")
	}
	return c, ""
}
func (c *context) pending() bool {
	var stat unix.Stat_t
	err := unix.Fstatat(int(c.home.Fd()), "import-pending", &stat, unix.AT_SYMLINK_NOFOLLOW)
	if err == unix.ENOENT {
		return true
	}
	if err != nil {
		return false
	}
	pending, ok := c.directory(c.home, "import-pending", filepath.Join(c.home.Name(), "import-pending"))
	if !ok || !c.private(c.guards[pending.Name()], 0700) {
		return false
	}
	names, ok := entries(pending, 0)
	return ok && len(names) == 0
}
func parseRegistration(data []byte, key string) (registration, bool) {
	zero := registration{}
	if !utf8.Valid(data) || !validJSONUnicode(string(data)) {
		return zero, false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return zero, false
	}
	fields := map[string]string{}
	version := false
	for decoder.More() {
		token, err := decoder.Token()
		name, ok := token.(string)
		if err != nil || !ok {
			return zero, false
		}
		if _, exists := fields[name]; exists {
			return zero, false
		}
		value, err := decoder.Token()
		if err != nil {
			return zero, false
		}
		if name == "version" {
			number, ok := value.(json.Number)
			if !ok || number.String() != "1" || version {
				return zero, false
			}
			version = true
			fields[name] = "1"
			continue
		}
		text, ok := value.(string)
		if !ok {
			return zero, false
		}
		fields[name] = text
	}
	last, err := decoder.Token()
	if err != nil || last != json.Delim('}') {
		return zero, false
	}
	if _, err = decoder.Token(); err != io.EOF {
		return zero, false
	}
	required := []string{"version", "sessionKey", "planIdBase64", "account", "root", "rootDevice", "rootInode", "record", "preparation", "planSHA256", "transaction"}
	if len(fields) != len(required) {
		return zero, false
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return zero, false
		}
	}
	encoded := fields["planIdBase64"]
	if len(encoded) > base64.StdEncoding.EncodedLen(1048576) {
		return zero, false
	}
	plan, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(plan) == 0 || len(plan) > 1048576 || !utf8.Valid(plan) || base64.StdEncoding.EncodeToString(plan) != encoded {
		return zero, false
	}
	account := fields["account"]
	if !accountName(account) || !absolute(fields["root"]) || !relativePath(fields["record"]) || !relativePath(fields["preparation"]) || !hexKey(fields["planSHA256"]) || !hexKey(fields["transaction"]) || fields["sessionKey"] != key {
		return zero, false
	}
	device, ok := decimal(fields["rootDevice"])
	if !ok {
		return zero, false
	}
	inode, ok := decimal(fields["rootInode"])
	if !ok {
		return zero, false
	}
	registration := registration{key: key, planID: string(plan), account: account, root: fields["root"], record: fields["record"], preparation: fields["preparation"], device: device, inode: inode}
	hash := sha256.New()
	_, _ = hash.Write([]byte("onboard-kit/session-key/v1\x00"))
	var length [8]byte
	for _, text := range []string{registration.planID, registration.account, registration.root} {
		binary.BigEndian.PutUint64(length[:], uint64(len(text)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write([]byte(text))
	}
	if hex.EncodeToString(hash.Sum(nil)) != key {
		return zero, false
	}
	return registration, true
}
func decimal(s string) (uint64, bool) {
	if s == "" || len(s) > 1 && s[0] == '0' {
		return 0, false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseUint(s, 10, 64)
	return value, err == nil
}
func accountName(s string) bool {
	if s == "" || len(s) > 4096 {
		return false
	}
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if i == 0 || c != '.' && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
func (c *context) same(next *context) bool {
	return reflect.DeepEqual(c.selected, next.selected) && reflect.DeepEqual(c.guards, next.guards) && reflect.DeepEqual(c.metadata, next.metadata) && reflect.DeepEqual(c.absences, next.absences)
}
func (c *context) load() (reviewcontract.Source, string) {
	next, code := openContext(c.args)
	if code != "" {
		return reviewcontract.Source{}, code
	}
	defer next.close()
	if !c.same(next) {
		return reviewcontract.Source{}, "context"
	}
	if !c.lock() {
		return reviewcontract.Source{}, "context"
	}
	defer c.unlock()
	if !c.pending() {
		return reviewcontract.Source{}, "pending"
	}
	if !c.absencesFresh() {
		return reviewcontract.Source{}, "context"
	}
	for _, s := range []*snapshot{&c.plan, &c.preparation, &c.record} {
		body, code := readBoundedBody(s.file, 1048576)
		if code != "" {
			return reviewcontract.Source{}, code
		}
		after, ok := guard(s.file)
		if !ok || !reflect.DeepEqual(after, s.guard) {
			return reviewcontract.Source{}, "context"
		}
		s.body = body
	}
	record, preparation, ok := locate(c.plan.body, c.selected.account)
	if !ok {
		return reviewcontract.Source{}, "review"
	}
	if record != c.selected.record || preparation != c.selected.preparation {
		return reviewcontract.Source{}, "context"
	}
	c.source = boundSource(c.plan.body, c.selected.account, c.selected.planID, c.preparation.body, c.record.body)
	return c.source, ""
}
func (c *context) fresh() string {
	if !c.unlock() {
		return "context"
	}
	next, code := openContext(c.args)
	if code != "" {
		return code
	}
	defer next.close()
	if !c.same(next) {
		return "context"
	}
	if len(c.locked) == 0 && !c.lock() {
		return "context"
	}
	if !c.pending() {
		return "pending"
	}
	if !c.absencesFresh() {
		return "context"
	}
	for _, pair := range [][2]*snapshot{{&c.plan, &next.plan}, {&c.preparation, &next.preparation}, {&c.record, &next.record}} {
		body, ok := boundedBody(pair[1].file, 1048576)
		if !ok || !bytes.Equal(body, pair[0].body) {
			return "context"
		}
		after, ok := guard(pair[1].file)
		if !ok || !reflect.DeepEqual(after, pair[0].guard) {
			return "context"
		}
	}
	if !c.stageIdle() {
		return "context"
	}
	return ""
}
func (c *context) stageName() string { return ".obk-record-edit-" + c.args.key }
func (c *context) stageIdle() bool {
	var stat unix.Stat_t
	err := unix.Fstatat(int(c.record.parent.Fd()), c.stageName(), &stat, unix.AT_SYMLINK_NOFOLLOW)
	if err == unix.ENOENT {
		return true
	}
	if err != nil {
		return false
	}
	fd, err := unix.Openat(int(c.record.parent.Fd()), c.stageName(), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	f := os.NewFile(uintptr(fd), c.stageName())
	defer f.Close()
	g, ok := guard(f)
	if !ok || !c.private(g, 0700) {
		return false
	}
	names, ok := entries(f, 0)
	return ok && len(names) == 0
}
