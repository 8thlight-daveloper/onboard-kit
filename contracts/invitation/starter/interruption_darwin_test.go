package starter_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

type stoppedOpening struct {
	phase, token string
	pending      *pending
	raw          []byte
}

type observationEvidence struct {
	profile      string
	initial      map[string]guard
	initialBytes map[string][]byte
}

type observedRoleFact uint8

const (
	roleIndeterminate observedRoleFact = iota
	rolePreimage
	rolePublished
	roleRetained
)

type observationLineage struct {
	root, home, namespace guard
	token, pendingHash    string
	roles                 [4]guard
	allPublished          bool
	sawReceipt            bool
	pendingGuard          guard
	pendingRaw            []byte
	payload               *pending
	ownerGuard            guard
	ownerRaw              []byte
	claimsGuard           guard
	claimsRaw             []byte
	parents               map[string]guard
}

type observedInvariant struct {
	violation, gap string
	facts          [4]observedRoleFact
}

func forbiddenRoleOrder(facts [4]observedRoleFact) bool {
	for i, fact := range facts {
		if fact != rolePublished {
			continue
		}
		for _, earlier := range facts[:i] {
			if earlier == rolePreimage {
				return true
			}
		}
	}
	return false
}

func canonicalObservation(body []byte, value any) bool {
	if json.Unmarshal(body, value) != nil {
		return false
	}
	canonical, err := json.Marshal(value)
	return err == nil && bytes.Equal(append(canonical, '\n'), body)
}

func stableObservedFile(t *testing.T, path string, limit int) ([]byte, guard, bool) {
	t.Helper()
	before, ok := observedGuard(t, path, "regular")
	if !ok {
		return nil, guard{}, false
	}
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, guard{}, false
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > int64(limit) {
		_ = f.Close()
		return nil, guard{}, false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || fmt.Sprint(stat.Dev) != before.Device || fmt.Sprint(stat.Ino) != before.Inode {
		_ = f.Close()
		return nil, guard{}, false
	}
	body, readError := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	afterInfo, statError := f.Stat()
	closeError := f.Close()
	after, present := observedGuard(t, path, "regular")
	if readError != nil || statError != nil || closeError != nil || !present || after != before || len(body) > limit || int64(len(body)) != info.Size() || !os.SameFile(info, afterInfo) || info.Size() != afterInfo.Size() {
		return nil, guard{}, false
	}
	return body, before, true
}

func definitelyAbsent(path string) bool {
	_, err := os.Lstat(path)
	return os.IsNotExist(err)
}

func validObservationSources(t *testing.T, s sandbox, e observationEvidence, p pending, raw []byte, o owner, root, home, namespace, pendingGuard guard) bool {
	t.Helper()
	if p.Version != 1 || o.Version != 1 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(p.Transaction) || o.Transaction != p.Transaction || o.RootDevice != root.Device || o.RootInode != root.Inode || o.Namespace != namespace || o.PendingSHA256 != digest(raw) || p.SessionKey != key(s.root) || p.PlanIDBase64 != "cGxhbi1h" || p.Account != "jules" || p.Root != s.root || p.Home != s.home || p.RootGuard != root || p.HomeGuard != home || p.Record != "records/jules/record.obk.yaml" || p.Preparation != "records/jules/preparation.obk.yaml" || len(p.Roles) != 4 || p.SessionDisplayBase64 != encoded(fixture(t, "session.txt")) {
		return false
	}
	if o.PendingTemporary != nil && *o.PendingTemporary != pendingGuard {
		return false
	}
	event := string(openedAt)
	if e.profile == "retained-record" {
		event = ""
	}
	if p.FirstOpenedAt != event {
		return false
	}
	paths, bodies, actions := observationArtifacts(t, s, e, p.Transaction)
	for i, r := range p.Roles {
		if r.Role != []string{"local-plan", "preparation", "session-metadata", "record"}[i] || r.Action != actions[i] || r.AfterBase64 != encoded(bodies[i]) || r.AfterSHA256 != digest(bodies[i]) || len(r.Parents) != 2 {
			return false
		}
		original, present := e.initial[paths[i]]
		if present != r.BeforePresent {
			return false
		}
		if present {
			if r.BeforeGuard == nil || *r.BeforeGuard != original || r.BeforeBase64 != encoded(e.initialBytes[paths[i]]) || r.BeforeSHA256 != digest(e.initialBytes[paths[i]]) {
				return false
			}
		} else if r.BeforeGuard != nil || r.BeforeBase64 != "" || r.BeforeSHA256 != "" {
			return false
		}
		base, parts := observationParentPath(s, i)
		for j, part := range parts {
			base = filepath.Join(base, part)
			original, existed := e.initial[base]
			parent := r.Parents[j]
			if parent.Ordinal != fmt.Sprint(j) || parent.Present != existed || existed && (parent.Guard == nil || *parent.Guard != original) || !existed && parent.Guard != nil {
				return false
			}
		}
	}
	return true
}

func observationParentPath(s sandbox, i int) (string, []string) {
	if i == 1 || i == 3 {
		return s.root, []string{"records", "jules"}
	}
	return s.home, []string{"sessions", key(s.root)}
}

func observedCreationClaims(t *testing.T, s sandbox, e observationEvidence, claims publication) map[string]guard {
	t.Helper()
	created := map[string]guard{}
	invalid := map[string]bool{}
	for i, role := range claims.Roles {
		base, parts := observationParentPath(s, i)
		last := -1
		for _, entry := range role.NewParents {
			ordinal, err := strconv.Atoi(entry.Ordinal)
			if err != nil || ordinal < 0 || ordinal >= len(parts) || entry.Ordinal != fmt.Sprint(ordinal) || ordinal <= last {
				continue
			}
			last = ordinal
			path := base
			for _, part := range parts[:ordinal+1] {
				path = filepath.Join(path, part)
			}
			g, ok := observedGuard(t, path, "directory")
			_, existed := e.initial[path]
			_, duplicate := created[path]
			if !ok || existed || duplicate || g != entry.Guard || !observerPrivate(g, "directory") {
				invalid[path] = true
				continue
			}
			created[path] = g
		}
	}
	for path := range invalid {
		delete(created, path)
	}
	return created
}

func observeRoleFacts(t *testing.T, s sandbox, e observationEvidence, p pending, claims *publication) ([4]observedRoleFact, [4]guard) {
	t.Helper()
	var facts [4]observedRoleFact
	var guards [4]guard
	created := map[string]guard{}
	if claims != nil {
		created = observedCreationClaims(t, s, e, *claims)
	}
	paths, bodies, _ := observationArtifacts(t, s, e, p.Transaction)
	for i, r := range p.Roles {
		facts[i], guards[i] = func() (observedRoleFact, guard) {
			base, parts := observationParentPath(s, i)
			for _, part := range parts {
				base = filepath.Join(base, part)
				original, existed := e.initial[base]
				current, present := observedGuard(t, base, "directory")
				if existed {
					if !present || current != original {
						return roleIndeterminate, guard{}
					}
				} else if present {
					claimed, ok := created[base]
					if !ok || current != claimed {
						return roleIndeterminate, guard{}
					}
				} else if !definitelyAbsent(base) {
					return roleIndeterminate, guard{}
				}
			}
			original, existed := e.initial[paths[i]]
			body, current, present := stableObservedFile(t, paths[i], 2097152)
			if r.Action == "retain" {
				if existed && present && current == original && bytes.Equal(body, e.initialBytes[paths[i]]) && bytes.Equal(body, bodies[i]) && (claims == nil || claims.Roles[i].Staging == nil && claims.Roles[i].Temporary == nil) {
					return roleRetained, current
				}
				return roleIndeterminate, guard{}
			}
			if existed && present && current == original && bytes.Equal(body, e.initialBytes[paths[i]]) || !existed && definitelyAbsent(paths[i]) {
				return rolePreimage, current
			}
			if !present || !bytes.Equal(body, bodies[i]) || claims == nil {
				return roleIndeterminate, guard{}
			}
			claim := claims.Roles[i]
			if claim.Staging == nil || claim.Temporary == nil || !observerPrivate(*claim.Staging, "directory") || !observerPrivate(*claim.Temporary, "regular") || current.Device != claim.Temporary.Device || current.Inode != claim.Temporary.Inode {
				return roleIndeterminate, guard{}
			}
			if existed && current.Security != original.Security || !existed && !observerPrivate(current, "regular") {
				return roleIndeterminate, guard{}
			}
			stage := filepath.Join(filepath.Dir(paths[i]), ".obk-start-"+p.Transaction+"-"+r.Role)
			stageGuard, ok := observedGuard(t, stage, "directory")
			if !ok || stageGuard != *claim.Staging || !definitelyAbsent(filepath.Join(stage, "payload")) {
				return roleIndeterminate, guard{}
			}
			entries, err := os.ReadDir(stage)
			if err != nil || len(entries) != 0 {
				return roleIndeterminate, guard{}
			}
			return rolePublished, current
		}()
	}
	return facts, guards
}

func allObservedRoles(facts [4]observedRoleFact) bool {
	for _, fact := range facts {
		if fact != rolePublished && fact != roleRetained {
			return false
		}
	}
	return true
}

func matchesSavedPublication(t *testing.T, s sandbox, e observationEvidence, history observationLineage) bool {
	t.Helper()
	if !history.allPublished || history.payload == nil {
		return false
	}
	for path, original := range history.parents {
		current, ok := observedGuard(t, path, "directory")
		if !ok || current != original {
			return false
		}
	}
	paths, bodies, _ := observationArtifacts(t, s, e, history.token)
	for i, path := range paths {
		body, g, ok := stableObservedFile(t, path, 2097152)
		if !ok || g != history.roles[i] || !bytes.Equal(body, bodies[i]) {
			return false
		}
	}
	return true
}

func sameObservationLineage(history observationLineage, root, home, namespace guard, token, pendingHash string) bool {
	return history.token == "" || history.root == root && history.home == home && history.namespace == namespace && history.token == token && history.pendingHash == pendingHash
}

func observedReceiptProof(t *testing.T, s sandbox, e observationEvidence, receipt complete, p pending, raw []byte, root, home, namespace guard, facts [4]observedRoleFact, guards [4]guard) (bool, bool) {
	t.Helper()
	securityRaw, err := json.Marshal(namespace.Security)
	if err != nil {
		return false, false
	}
	if receipt.Version != 1 || receipt.Transaction != p.Transaction || receipt.SessionKey != key(s.root) || receipt.PendingSHA256 != digest(raw) || receipt.RootDevice != root.Device || receipt.RootInode != root.Inode || receipt.HomeDevice != home.Device || receipt.HomeInode != home.Inode || receipt.NamespaceDevice != namespace.Device || receipt.NamespaceInode != namespace.Inode || receipt.NamespaceSecuritySHA256 != digest(securityRaw) || len(receipt.Roles) != 4 {
		return true, false
	}
	_, bodies, _ := observationArtifacts(t, s, e, p.Transaction)
	verified := true
	for i, role := range receipt.Roles {
		if role.Role != p.Roles[i].Role || role.SHA256 != digest(bodies[i]) {
			return true, false
		}
		switch facts[i] {
		case rolePreimage:
			return true, false
		case roleIndeterminate:
			verified = false
		case rolePublished, roleRetained:
			securityRaw, err := json.Marshal(guards[i].Security)
			if err != nil {
				return false, false
			}
			if role.Device != guards[i].Device || role.Inode != guards[i].Inode || role.SecuritySHA256 != digest(securityRaw) {
				return true, false
			}
		}
	}
	return false, verified
}

func observeNativeInvariants(t *testing.T, s sandbox, e observationEvidence, history *observationLineage) observedInvariant {
	t.Helper()
	root, rootOK := observedGuard(t, s.root, "directory")
	home, homeOK := observedGuard(t, s.home, "directory")
	namespace, namespaceOK := observedGuard(t, s.pending(), "directory")
	if !rootOK || root != e.initial[s.root] || !homeOK || !observerPrivate(home, "directory") || !namespaceOK || !observerPrivate(namespace, "directory") {
		return observedInvariant{gap: "native-root-home-namespace-unproved"}
	}
	if original, existed := e.initial[s.home]; existed && home != original {
		return observedInvariant{gap: "original-home-changed"}
	}
	if history.token != "" && (history.root != root || history.home != home || history.namespace != namespace) {
		return observedInvariant{gap: "native-lineage-changed"}
	}
	pendingPath, receiptPath := filepath.Join(s.pending(), "pending.json"), filepath.Join(s.pending(), "complete.json")
	ownerRaw, ownerGuard, ownerOK := stableObservedFile(t, filepath.Join(s.pending(), "owner.json"), 16384)
	claimRaw, claimGuard, claimOK := stableObservedFile(t, filepath.Join(s.pending(), "publication.json"), 16384)
	if history.token != "" && definitelyAbsent(receiptPath) {
		pendingRaw, pendingGuard, pendingOK := stableObservedFile(t, pendingPath, 33554432)
		if history.sawReceipt && pendingOK && pendingGuard == history.pendingGuard && bytes.Equal(pendingRaw, history.pendingRaw) {
			return observedInvariant{violation: "completion-proof-lost-with-personal-pending"}
		}
		ownedRemainder := ownerOK && ownerGuard == history.ownerGuard && bytes.Equal(ownerRaw, history.ownerRaw) || claimOK && claimGuard == history.claimsGuard && bytes.Equal(claimRaw, history.claimsRaw)
		if definitelyAbsent(pendingPath) && ownedRemainder && matchesSavedPublication(t, s, e, *history) {
			return observedInvariant{violation: "published-cleanup-lost-pending-and-receipt"}
		}
	}
	var p pending
	var o owner
	raw, pendingGuard, hasPending := stableObservedFile(t, pendingPath, 33554432)
	if !hasPending || !canonicalObservation(raw, &p) || !ownerOK || !observerPrivate(ownerGuard, "regular") || !canonicalObservation(ownerRaw, &o) || !observerPrivate(pendingGuard, "regular") || !validObservationSources(t, s, e, p, raw, o, root, home, namespace, pendingGuard) {
		return observedInvariant{gap: "complete-pending-lineage-unproved"}
	}
	if !sameObservationLineage(*history, root, home, namespace, p.Transaction, digest(raw)) {
		return observedInvariant{gap: "transaction-or-pending-lineage-changed"}
	}
	var claims *publication
	var claim publication
	if claimOK && observerPrivate(claimGuard, "regular") && canonicalObservation(claimRaw, &claim) && claim.Version == 1 && claim.Transaction == p.Transaction && claim.PendingSHA256 == digest(raw) && len(claim.Roles) == 4 {
		valid := true
		for i, role := range claim.Roles {
			valid = valid && role.Role == p.Roles[i].Role && role.NewParents != nil
		}
		if valid {
			claims = &claim
		}
	}
	facts, guards := observeRoleFacts(t, s, e, p, claims)
	history.root, history.home, history.namespace = root, home, namespace
	history.token, history.pendingHash = p.Transaction, digest(raw)
	history.pendingGuard, history.pendingRaw, history.payload = pendingGuard, raw, &p
	history.ownerGuard, history.ownerRaw = ownerGuard, ownerRaw
	if forbiddenRoleOrder(facts) {
		return observedInvariant{violation: "forbidden-published-role-subset", facts: facts}
	}
	if claims != nil {
		history.claimsGuard, history.claimsRaw = claimGuard, claimRaw
	}
	if allObservedRoles(facts) {
		history.allPublished, history.roles = true, guards
		history.parents = map[string]guard{}
		for i := range p.Roles {
			base, parts := observationParentPath(s, i)
			for _, part := range parts {
				base = filepath.Join(base, part)
				history.parents[base] = objectGuard(t, base)
			}
		}
	}
	receiptRaw, receiptGuard, receiptOK := stableObservedFile(t, receiptPath, 16384)
	if !receiptOK {
		return observedInvariant{facts: facts, gap: "no-stable-final-receipt-observed"}
	}
	if !observerPrivate(receiptGuard, "regular") {
		return observedInvariant{facts: facts, gap: "final-receipt-native-controls-unproved"}
	}
	var receipt complete
	if !canonicalObservation(receiptRaw, &receipt) {
		return observedInvariant{violation: "attributable-final-receipt-malformed", facts: facts}
	}
	invalid, verified := observedReceiptProof(t, s, e, receipt, p, raw, root, home, namespace, facts, guards)
	if invalid {
		return observedInvariant{violation: "attributable-final-receipt-invalid", facts: facts}
	}
	if !verified {
		return observedInvariant{facts: facts, gap: "final-receipt-role-native-proof-incomplete"}
	}
	history.sawReceipt = true
	return observedInvariant{facts: facts}
}

func captureObservation(t *testing.T, s sandbox, profile string) observationEvidence {
	t.Helper()
	e := observationEvidence{profile: profile, initial: map[string]guard{}, initialBytes: map[string][]byte{}}
	for _, path := range []string{s.root, s.home, filepath.Join(s.root, "records"), filepath.Dir(s.record()), s.record(), s.prep(), filepath.Join(s.home, "sessions"), s.registration(), filepath.Join(s.registration(), "plan.obk.yaml"), filepath.Join(s.registration(), "session.json")} {
		if info, err := os.Lstat(path); err == nil {
			e.initial[path] = objectGuard(t, path)
			if info.Mode().IsRegular() {
				body, ok := observedBody(path, 2097152)
				if !ok {
					t.Fatal("cannot freeze bounded native source evidence")
				}
				e.initialBytes[path] = body
			}
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return e
}

func observerPrivate(g guard, kind string) bool {
	mode := "0600"
	if kind == "directory" {
		mode = "0700"
	}
	return g.Kind == kind && g.Security.Mode == mode && g.Security.UID == fmt.Sprint(os.Geteuid()) && g.Security.GID == fmt.Sprint(os.Getegid()) && !g.Security.AccessPresent && !g.Security.DefaultPresent && g.Security.AccessBase64 == "" && g.Security.DefaultBase64 == ""
}

func observedGuard(t *testing.T, path, kind string) (guard, bool) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		return guard{}, false
	}
	if kind == "directory" && !info.IsDir() || kind == "regular" && !info.Mode().IsRegular() {
		return guard{}, false
	}
	return objectGuard(t, path), true
}

func observedBody(path string, limit int) ([]byte, bool) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > int64(limit) {
		return nil, false
	}
	body, err := os.ReadFile(path)
	return body, err == nil && len(body) <= limit
}

func observationArtifacts(t *testing.T, s sandbox, e observationEvidence, transaction string) ([]string, [][]byte, []string) {
	t.Helper()
	paths := []string{filepath.Join(s.registration(), "plan.obk.yaml"), s.prep(), filepath.Join(s.registration(), "session.json"), s.record()}
	metadata := metadataBytes(t, s)
	metadata = bytes.ReplaceAll(metadata, []byte(recoveryToken), []byte(transaction))
	bodies := [][]byte{fixture(t, "plan.yaml"), fixture(t, "preparation.yaml"), metadata, fixture(t, "begun.yaml")}
	actions := []string{"install", "install", "install", "create"}
	if e.profile == "reconcile" {
		actions[3] = "reconcile"
	}
	if e.profile == "retained-record" {
		actions[3] = "retain"
		bodies[3] = fixture(t, "existing.yaml")
	}
	return paths, bodies, actions
}

func validObservedReceipt(t *testing.T, s sandbox, e observationEvidence, c complete, raw []byte, p *pending) bool {
	t.Helper()
	token := regexp.MustCompile(`^[0-9a-f]{64}$`)
	hash := regexp.MustCompile(`^[0-9a-f]{64}$`)
	root, home, namespace := objectGuard(t, s.root), objectGuard(t, s.home), objectGuard(t, s.pending())
	securityBytes, err := json.Marshal(namespace.Security)
	if err != nil {
		return false
	}
	if c.Version != 1 || !token.MatchString(c.Transaction) || !hash.MatchString(c.PendingSHA256) || c.SessionKey != key(s.root) || c.RootDevice != root.Device || c.RootInode != root.Inode || root != e.initial[s.root] || c.HomeDevice != home.Device || c.HomeInode != home.Inode || !observerPrivate(home, "directory") || c.NamespaceDevice != namespace.Device || c.NamespaceInode != namespace.Inode || c.NamespaceSecuritySHA256 != digest(securityBytes) || !observerPrivate(namespace, "directory") || len(c.Roles) != 4 {
		return false
	}
	if p != nil && (c.PendingSHA256 != digest(raw) || c.Transaction != p.Transaction) {
		return false
	}
	paths, bodies, _ := observationArtifacts(t, s, e, c.Transaction)
	for i, name := range []string{"local-plan", "preparation", "session-metadata", "record"} {
		g, ok := observedGuard(t, paths[i], "regular")
		if !ok {
			return false
		}
		body, ok := observedBody(paths[i], 1048576)
		if i == 2 {
			body, ok = observedBody(paths[i], 2097152)
		}
		if !ok || !bytes.Equal(body, bodies[i]) {
			return false
		}
		securityBytes, err := json.Marshal(g.Security)
		if err != nil {
			return false
		}
		r := c.Roles[i]
		if r.Role != name || r.SHA256 != digest(bodies[i]) || r.Device != g.Device || r.Inode != g.Inode || r.SecuritySHA256 != digest(securityBytes) {
			return false
		}
		if before, present := e.initial[paths[i]]; present {
			if g.Security != before.Security {
				return false
			}
		} else if !observerPrivate(g, "regular") {
			return false
		}
	}
	return true
}

func validObservedPending(t *testing.T, s sandbox, e observationEvidence, p pending, raw []byte, o owner, claims *publication) (int, bool, bool) {
	t.Helper()
	root, home := objectGuard(t, s.root), objectGuard(t, s.home)
	if p.Version != 1 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(p.Transaction) || p.Transaction != o.Transaction || o.PendingSHA256 != digest(raw) || p.SessionKey != key(s.root) || p.PlanIDBase64 != encoded([]byte("plan-a")) || p.Account != "jules" || p.Root != s.root || p.Home != s.home || p.Record != "records/jules/record.obk.yaml" || p.Preparation != "records/jules/preparation.obk.yaml" || p.RootGuard != root || root != e.initial[s.root] || p.HomeGuard != home || !observerPrivate(home, "directory") || len(p.Roles) != 4 || p.SessionDisplayBase64 != encoded(fixture(t, "session.txt")) {
		return 0, false, false
	}
	event := string(openedAt)
	if e.profile == "retained-record" {
		event = ""
	}
	if p.FirstOpenedAt != event {
		return 0, false, false
	}
	if claims != nil && (claims.Version != 1 || claims.Transaction != p.Transaction || claims.PendingSHA256 != digest(raw) || len(claims.Roles) != 4) {
		return 0, false, false
	}
	if o.PendingTemporary != nil {
		g, ok := observedGuard(t, filepath.Join(s.pending(), "pending.json"), "regular")
		if !ok || g != *o.PendingTemporary {
			return 0, false, false
		}
	}
	paths, bodies, actions := observationArtifacts(t, s, e, p.Transaction)
	created := map[string]guard{}
	if claims != nil {
		for i, c := range claims.Roles {
			if c.Role != []string{"local-plan", "preparation", "session-metadata", "record"}[i] || c.NewParents == nil {
				return 0, false, false
			}
			base := s.home
			parts := []string{"sessions", key(s.root)}
			if i == 1 || i == 3 {
				base = s.root
				parts = []string{"records", "jules"}
			}
			prior := -1
			for _, entry := range c.NewParents {
				index, err := strconv.Atoi(entry.Ordinal)
				if err != nil || index < 0 || index >= len(parts) || fmt.Sprint(index) != entry.Ordinal || index <= prior {
					return 0, false, false
				}
				prior = index
				path := base
				for j := 0; j <= index; j++ {
					path = filepath.Join(path, parts[j])
				}
				if _, existed := e.initial[path]; existed {
					return 0, false, false
				}
				if _, duplicate := created[path]; duplicate {
					return 0, false, false
				}
				g, ok := observedGuard(t, path, "directory")
				if !ok || g != entry.Guard || !observerPrivate(g, "directory") {
					return 0, false, false
				}
				created[path] = g
			}
		}
	}
	prefix := 0
	blocked := false
	finalPayload := false
	for i, r := range p.Roles {
		if r.Role != []string{"local-plan", "preparation", "session-metadata", "record"}[i] || r.Action != actions[i] || r.AfterBase64 != encoded(bodies[i]) || r.AfterSHA256 != digest(bodies[i]) || len(r.Parents) != 2 {
			return 0, false, false
		}
		before, present := e.initial[paths[i]]
		if r.BeforePresent != present {
			return 0, false, false
		}
		if present {
			beforeBytes := fixture(t, "existing.yaml")
			if e.profile == "reconcile" {
				beforeBytes = bytes.ReplaceAll(fixture(t, "begun.yaml"), []byte(rawPlanHash), []byte(strings.Repeat("0", 64)))
			}
			if r.BeforeGuard == nil || *r.BeforeGuard != before || r.BeforeBase64 != encoded(beforeBytes) || r.BeforeSHA256 != digest(beforeBytes) {
				return 0, false, false
			}
		} else if r.BeforeGuard != nil || r.BeforeBase64 != "" || r.BeforeSHA256 != "" {
			return 0, false, false
		}
		base := s.home
		parts := []string{"sessions", key(s.root)}
		if i == 1 || i == 3 {
			base = s.root
			parts = []string{"records", "jules"}
		}
		for j, part := range parts {
			base = filepath.Join(base, part)
			parent := r.Parents[j]
			original, existed := e.initial[base]
			if parent.Ordinal != fmt.Sprint(j) || parent.Present != existed {
				return 0, false, false
			}
			if existed {
				g, ok := observedGuard(t, base, "directory")
				if !ok || parent.Guard == nil || *parent.Guard != original || g != original {
					return 0, false, false
				}
			} else {
				if parent.Guard != nil {
					return 0, false, false
				}
				if g, ok := observedGuard(t, base, "directory"); ok {
					if claimed, ok := created[base]; !ok || g != claimed {
						return 0, false, false
					}
				} else if _, err := os.Lstat(base); !os.IsNotExist(err) {
					return 0, false, false
				}
			}
		}
		var claim *claimRole
		if claims != nil {
			claim = &claims.Roles[i]
		}
		stage := filepath.Join(filepath.Dir(paths[i]), ".obk-start-"+p.Transaction+"-"+r.Role)
		payload := filepath.Join(stage, "payload")
		final := security{}
		if present {
			final = before.Security
		}
		if claim != nil && claim.Staging != nil {
			g, ok := observedGuard(t, stage, "directory")
			if !ok || g != *claim.Staging || !observerPrivate(g, "directory") || claim.Temporary == nil || !observerPrivate(*claim.Temporary, "regular") {
				return 0, false, false
			}
			entries, err := os.ReadDir(stage)
			if err != nil || len(entries) > 1 || len(entries) == 1 && entries[0].Name() != "payload" {
				return 0, false, false
			}
			if g, ok := observedGuard(t, payload, "regular"); ok {
				if g.Device != claim.Temporary.Device || g.Inode != claim.Temporary.Inode || g.Security != claim.Temporary.Security && (!present || g.Security != final) {
					return 0, false, false
				}
				body, ok := observedBody(payload, 2097152)
				if !ok || !bytes.HasPrefix(bodies[i], body) {
					return 0, false, false
				}
				if present && g.Security == final && g.Security != claim.Temporary.Security {
					finalPayload = true
				}
			} else if _, err := os.Lstat(payload); !os.IsNotExist(err) {
				return 0, false, false
			}
		} else {
			if claim != nil && claim.Temporary != nil {
				return 0, false, false
			}
			if _, err := os.Lstat(stage); !os.IsNotExist(err) {
				return 0, false, false
			}
		}
		body, exists := observedBody(paths[i], 2097152)
		applied := false
		if actions[i] == "retain" {
			g, ok := observedGuard(t, paths[i], "regular")
			if !exists || !bytes.Equal(body, bodies[i]) || !ok || g != before || claim != nil && (claim.Staging != nil || claim.Temporary != nil) {
				return 0, false, false
			}
			applied = true
		} else if exists && bytes.Equal(body, bodies[i]) {
			g, ok := observedGuard(t, paths[i], "regular")
			if !ok || claim == nil || claim.Temporary == nil || g.Device != claim.Temporary.Device || g.Inode != claim.Temporary.Inode || present && g.Security != final || !present && !observerPrivate(g, "regular") {
				return 0, false, false
			}
			applied = true
		} else if present {
			g, ok := observedGuard(t, paths[i], "regular")
			expected, err := base64.StdEncoding.DecodeString(r.BeforeBase64)
			if err != nil || !exists || !bytes.Equal(body, expected) || !ok || g != before {
				return 0, false, false
			}
		} else if _, err := os.Lstat(paths[i]); !os.IsNotExist(err) {
			return 0, false, false
		}
		if claim != nil && claim.Temporary != nil && !applied {
			if _, ok := observedGuard(t, payload, "regular"); !ok {
				return 0, false, false
			}
		}
		if applied {
			if blocked && actions[i] != "retain" {
				return 0, false, false
			}
			if !blocked {
				prefix = i + 1
			}
		} else {
			blocked = true
		}
	}
	return prefix, finalPayload, true
}

func readProtocol(path string, limit int, value any) ([]byte, bool) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > int64(limit) {
		return nil, false
	}
	body, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil || len(body) > limit {
		return nil, false
	}
	if json.Unmarshal(body, value) != nil {
		return nil, false
	}
	canonical, err := json.Marshal(value)
	if err != nil || !bytes.Equal(append(canonical, '\n'), body) {
		return nil, false
	}
	return body, true
}

func observeStoppedOpening(t *testing.T, s sandbox, evidence observationEvidence) stoppedOpening {
	t.Helper()
	entries, err := os.ReadDir(s.pending())
	if os.IsNotExist(err) {
		return stoppedOpening{phase: "pre-namespace"}
	}
	if err != nil {
		return stoppedOpening{phase: "unclassified"}
	}
	if len(entries) == 0 {
		g := objectGuard(t, s.pending())
		if g.Security.Mode == "0700" && !g.Security.AccessPresent && !g.Security.DefaultPresent {
			return stoppedOpening{phase: "idle"}
		}
		return stoppedOpening{phase: "unclassified"}
	}
	for _, entry := range entries {
		switch entry.Name() {
		case "owner.json", "pending.json", "publication.json", "complete.json", ".owner.next", ".pending.next", ".publication.next", ".complete.next":
		default:
			return stoppedOpening{phase: "unclassified"}
		}
		g, ok := observedGuard(t, filepath.Join(s.pending(), entry.Name()), "regular")
		if !ok || !observerPrivate(g, "regular") {
			return stoppedOpening{phase: "unclassified"}
		}
	}
	var p pending
	raw, hasPending := readProtocol(filepath.Join(s.pending(), "pending.json"), 33554432, &p)
	var c complete
	_, hasReceipt := readProtocol(filepath.Join(s.pending(), "complete.json"), 16384, &c)
	root := objectGuard(t, s.root)
	if hasReceipt {
		if _, err := os.Lstat(filepath.Join(s.pending(), "pending.json")); os.IsNotExist(err) {
			if validObservedReceipt(t, s, evidence, c, nil, nil) {
				return stoppedOpening{phase: "receipt-only", token: c.Transaction}
			}
			return stoppedOpening{phase: "unclassified"}
		}
	}
	var o owner
	_, hasOwner := readProtocol(filepath.Join(s.pending(), "owner.json"), 16384, &o)
	if !hasOwner {
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || info.Size() != 0 {
				return stoppedOpening{phase: "unclassified"}
			}
		}
		return stoppedOpening{phase: "unclaimed-empty"}
	}
	if o.Version != 1 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(o.Transaction) || o.Namespace != objectGuard(t, s.pending()) || !observerPrivate(o.Namespace, "directory") || o.RootDevice != root.Device || o.RootInode != root.Inode {
		return stoppedOpening{phase: "unclassified"}
	}
	if !hasPending {
		if o.PendingSHA256 != "" && !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(o.PendingSHA256) {
			return stoppedOpening{phase: "unclassified"}
		}
		if o.PendingTemporary != nil {
			temporary := filepath.Join(s.pending(), ".pending.next")
			if g, ok := observedGuard(t, temporary, "regular"); ok && g == *o.PendingTemporary && observerPrivate(g, "regular") {
				for _, path := range []string{filepath.Join(s.registration(), "plan.obk.yaml"), s.prep(), filepath.Join(s.registration(), "session.json"), s.record()} {
					if initial, existed := evidence.initial[path]; existed {
						current, ok := observedGuard(t, path, "regular")
						body, hasBody := observedBody(path, 2097152)
						if !ok || current != initial || !hasBody || !bytes.Equal(body, evidence.initialBytes[path]) {
							return stoppedOpening{phase: "unclassified"}
						}
					} else if _, err := os.Lstat(path); !os.IsNotExist(err) {
						return stoppedOpening{phase: "unclassified"}
					}
				}
				for _, path := range []string{filepath.Join(s.root, "records"), filepath.Dir(s.record()), filepath.Join(s.home, "sessions"), s.registration()} {
					if original, existed := evidence.initial[path]; existed {
						current, ok := observedGuard(t, path, "directory")
						if !ok || current != original {
							return stoppedOpening{phase: "unclassified"}
						}
					} else if _, err := os.Lstat(path); !os.IsNotExist(err) {
						return stoppedOpening{phase: "unclassified"}
					}
				}
				return stoppedOpening{phase: "claimed-incomplete", token: o.Transaction}
			}
		}
		return stoppedOpening{phase: "unclassified"}
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".next") {
			return stoppedOpening{phase: "unclassified"}
		}
	}
	var claimData publication
	_, hasClaims := readProtocol(filepath.Join(s.pending(), "publication.json"), 16384, &claimData)
	var claims *publication
	if hasClaims {
		claims = &claimData
	} else if _, err := os.Lstat(filepath.Join(s.pending(), "publication.json")); !os.IsNotExist(err) {
		return stoppedOpening{phase: "unclassified"}
	}
	prefix, finalPayload, verified := validObservedPending(t, s, evidence, p, raw, o, claims)
	if !verified {
		return stoppedOpening{phase: "unclassified"}
	}
	if hasReceipt {
		if prefix == 4 && validObservedReceipt(t, s, evidence, c, raw, &p) {
			return stoppedOpening{phase: "receipt-pending", token: p.Transaction, pending: &p, raw: raw}
		}
		return stoppedOpening{phase: "unclassified"}
	}
	phase := []string{"pending-no-role", "after-local-plan", "after-preparation", "after-session-metadata", "all-roles-no-receipt"}[prefix]
	if finalPayload {
		phase = "payload-final-controls"
	}
	if evidence.profile == "retained-record" && prefix < 3 {
		phase = "retained-record/" + phase
	}
	return stoppedOpening{phase: phase, token: p.Transaction, pending: &p, raw: raw}
}

func TestDarwinObservedProcessDeath(t *testing.T) {
	if os.Getenv("OBK_STARTER_OBSERVE_DEATH") != "1" {
		t.Skip("native observed interruption layer not requested; phase coverage remains unverified")
	}
	desired := []string{"idle", "unclaimed-empty", "claimed-incomplete", "pending-no-role", "after-local-plan", "after-preparation", "after-session-metadata", "retained-record/pending-no-role", "retained-record/after-local-plan", "retained-record/after-preparation", "payload-final-controls", "all-roles-no-receipt", "receipt-pending", "receipt-only"}
	seen := map[string]bool{}
	gaps := map[string]int{}
	delays := []time.Duration{0, time.Microsecond, 5 * time.Microsecond, 25 * time.Microsecond, 100 * time.Microsecond, 500 * time.Microsecond, time.Millisecond, 2 * time.Millisecond, 5 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond, 50 * time.Millisecond}
	for _, profile := range []string{"fresh", "reconcile", "retained-record"} {
		for delayIndex, delay := range delays {
			s := setup(t)
			if profile == "reconcile" {
				put(t, s.record(), bytes.ReplaceAll(fixture(t, "begun.yaml"), []byte(rawPlanHash), []byte(strings.Repeat("0", 64))))
				if err := os.Chmod(s.record(), 0400); err != nil {
					t.Fatal(err)
				}
			}
			if profile == "retained-record" {
				put(t, s.record(), fixture(t, "existing.yaml"))
			}
			evidence := captureObservation(t, s, profile)
			c := launchChild(t, s)
			c.until(t, "Identity: choose confirm jules or cancel:")
			c.send(t, "confirm jules\n")
			c.until(t, "Review complete. Choose begin or cancel:")
			c.send(t, "begin\n")
			var state stoppedOpening
			history := observationLineage{}
			for observation := 0; observation < 8; observation++ {
				pause := delay
				if observation > 0 {
					pause = []time.Duration{0, time.Microsecond, 5 * time.Microsecond, 25 * time.Microsecond, 100 * time.Microsecond, 500 * time.Microsecond, time.Millisecond, 2 * time.Millisecond}[observation]
				}
				if pause > 0 {
					timer := time.NewTimer(pause)
					<-timer.C
				}
				if err := c.command.Process.Signal(syscall.SIGSTOP); err != nil {
					t.Fatal("cannot stop owned native helper")
				}
				var status syscall.WaitStatus
				var pid int
				var waitError error
				for attempt := 0; attempt < 4; attempt++ {
					pid, waitError = syscall.Wait4(c.command.Process.Pid, &status, syscall.WUNTRACED, nil)
					if waitError != syscall.EINTR {
						break
					}
				}
				if waitError != nil || pid != c.command.Process.Pid || status&0x7f != 0x7f || status>>8 != syscall.WaitStatus(syscall.SIGSTOP) {
					t.Fatal("cannot confirm owned helper stopped before classification")
				}
				invariant := observeNativeInvariants(t, s, evidence, &history)
				if invariant.violation != "" {
					if err := c.command.Process.Kill(); err != nil {
						t.Fatal("cannot kill owned helper after proved forbidden state")
					}
					_ = c.input.Close()
					_ = c.command.Wait()
					c.waited = true
					c.cancel()
					t.Fatalf("proved forbidden native state: %s profile=%s pid=%d sample=%d facts=%v token=%s pending=%s root=%s:%s home=%s:%s namespace=%s:%s", invariant.violation, profile, pid, observation, invariant.facts, history.token, history.pendingHash, history.root.Device, history.root.Inode, history.home.Device, history.home.Inode, history.namespace.Device, history.namespace.Inode)
				}
				if invariant.gap != "" {
					gaps[invariant.gap]++
				}
				state = observeStoppedOpening(t, s, evidence)
				t.Logf("native observation profile=%s attempt=%d sample=%d phase=%s facts=%v invariant-gap=%q", profile, delayIndex, observation, state.phase, invariant.facts, invariant.gap)
				if observation < 7 && delayIndex%3 == 0 && state.phase != "idle" && state.phase != "unclaimed-empty" && state.phase != "claimed-incomplete" {
					if err := c.command.Process.Signal(syscall.SIGCONT); err != nil {
						t.Fatal("cannot continue owned helper between bounded observations")
					}
					continue
				}
				break
			}
			if state.phase == "unclassified" || state.phase == "pre-namespace" || seen[state.phase] {
				if err := c.command.Process.Signal(syscall.SIGCONT); err != nil {
					t.Fatal(err)
				}
				c.send(t, "exit\n")
				c.finish(t)
				continue
			}
			if err := c.command.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			_ = c.input.Close()
			_ = c.command.Wait()
			c.waited = true
			c.cancel()
			seen[state.phase] = true
			t.Logf("observed real stopped/killed phase: %s", state.phase)
			if err := os.Remove(s.invitation); err != nil {
				t.Fatal(err)
			}
			clock := 0
			switch state.phase {
			case "unclaimed-empty":
				before := snapshot(t, s.base)
				got := invoke(t, newStarter(runtimeFacts(t, &clock)), resumeArgs(s), "clean\n")
				if got.exit != 1 {
					t.Error("unclaimed crash objects were adopted")
				}
				unchanged(t, s, before)
			case "claimed-incomplete":
				args := []string{"--resume-import", state.token, "--root", s.root, "--home", s.home}
				got := invoke(t, newStarter(runtimeFacts(t, &clock)), args, "clean\n")
				wantReport(t, got, 0, "cleaned", "start.staging-cleaned", "Owned incomplete staging was cleaned; reopen the invitation to begin.")
				assertIdle(t, s)
			case "idle":
				if entries, err := os.ReadDir(s.pending()); err != nil || len(entries) != 0 {
					t.Error("observed idle retained personal data")
				}
			default:
				args := []string{"--resume-import", state.token, "--root", s.root, "--home", s.home}
				action := "resume\nexit\n"
				if strings.HasPrefix(state.phase, "receipt") {
					action = "clean\nexit\n"
				}
				got := invoke(t, newStarter(runtimeFacts(t, &clock)), args, action)
				if state.phase == "receipt-only" {
					wantReport(t, got, 0, "cleaned", "start.cleaned", "Completed recovery data was cleaned; no session was displayed.")
				} else {
					wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
					if state.pending != nil {
						paths := []string{filepath.Join(s.registration(), "plan.obk.yaml"), s.prep(), filepath.Join(s.registration(), "session.json"), s.record()}
						for i, path := range paths {
							body, err := base64.StdEncoding.DecodeString(state.pending.Roles[i].AfterBase64)
							if err != nil {
								t.Fatal(err)
							}
							assertBytes(t, path, body)
						}
						display, err := base64.StdEncoding.DecodeString(state.pending.SessionDisplayBase64)
						if err != nil {
							t.Fatal(err)
						}
						if !strings.Contains(got.out, string(display)) {
							t.Error("real interruption recovery changed frozen display")
						}
					}
				}
				assertIdle(t, s)
			}
			if clock != 0 {
				t.Error("real interruption recovery sampled another event")
			}
		}
	}
	for _, phase := range desired {
		if !seen[phase] {
			t.Logf("UNOBSERVED native interruption phase gap: %s", phase)
		}
	}
	for _, reason := range []string{"native-root-home-namespace-unproved", "original-home-changed", "native-lineage-changed", "complete-pending-lineage-unproved", "transaction-or-pending-lineage-changed", "no-stable-final-receipt-observed", "final-receipt-native-controls-unproved", "final-receipt-role-native-proof-incomplete"} {
		if gaps[reason] > 0 {
			t.Logf("INCONCLUSIVE native invariant observations: %s count=%d", reason, gaps[reason])
		}
	}
	for _, witness := range []string{"forbidden-published-role-subset", "attributable-final-receipt-malformed", "attributable-final-receipt-invalid", "completion-proof-lost-with-personal-pending", "published-cleanup-lost-pending-and-receipt"} {
		t.Logf("UNREACHED forbidden-state witness is not equivalence or a mutant kill: %s", witness)
	}
	if len(seen) == 0 {
		t.Skip(fmt.Sprintf("no required persistent phase observed in %d finite attempts across three profiles", 3*len(delays)))
	}
}
