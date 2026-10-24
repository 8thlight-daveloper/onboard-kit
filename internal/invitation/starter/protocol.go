package starter

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type metadata struct {
	Version      int    `json:"version"`
	SessionKey   string `json:"sessionKey"`
	PlanIDBase64 string `json:"planIdBase64"`
	Account      string `json:"account"`
	Root         string `json:"root"`
	RootDevice   string `json:"rootDevice"`
	RootInode    string `json:"rootInode"`
	Record       string `json:"record"`
	Preparation  string `json:"preparation"`
	PlanSHA256   string `json:"planSHA256"`
	Transaction  string `json:"transaction"`
}
type ownerIntent struct {
	Version          int          `json:"version"`
	Transaction      string       `json:"transaction"`
	RootDevice       string       `json:"rootDevice"`
	RootInode        string       `json:"rootInode"`
	Namespace        objectGuard  `json:"namespace"`
	PendingSHA256    string       `json:"pendingSHA256"`
	PendingTemporary *objectGuard `json:"pendingTemporary"`
}
type pendingRole struct {
	Role          string        `json:"role"`
	Action        string        `json:"action"`
	BeforePresent bool          `json:"beforePresent"`
	BeforeBase64  string        `json:"beforeBase64"`
	BeforeSHA256  string        `json:"beforeSHA256"`
	BeforeGuard   *objectGuard  `json:"beforeGuard"`
	AfterBase64   string        `json:"afterBase64"`
	AfterSHA256   string        `json:"afterSHA256"`
	Parents       []parentGuard `json:"parents"`
}
type pending struct {
	Version              int           `json:"version"`
	Transaction          string        `json:"transaction"`
	SessionKey           string        `json:"sessionKey"`
	PlanIDBase64         string        `json:"planIdBase64"`
	Account              string        `json:"account"`
	Root                 string        `json:"root"`
	RootGuard            objectGuard   `json:"rootGuard"`
	Home                 string        `json:"home"`
	HomeGuard            objectGuard   `json:"homeGuard"`
	Record               string        `json:"record"`
	Preparation          string        `json:"preparation"`
	FirstOpenedAt        string        `json:"firstOpenedAt"`
	SessionDisplayBase64 string        `json:"sessionDisplayBase64"`
	Roles                []pendingRole `json:"roles"`
}
type createdParent struct {
	Ordinal string      `json:"ordinal"`
	Guard   objectGuard `json:"guard"`
}
type roleClaim struct {
	Role       string          `json:"role"`
	NewParents []createdParent `json:"newParents"`
	Staging    *objectGuard    `json:"staging"`
	Temporary  *objectGuard    `json:"temporary"`
}
type publicationClaim struct {
	Version       int         `json:"version"`
	Transaction   string      `json:"transaction"`
	PendingSHA256 string      `json:"pendingSHA256"`
	Roles         []roleClaim `json:"roles"`
}
type receiptRole struct {
	Role           string `json:"role"`
	SHA256         string `json:"sha256"`
	Device         string `json:"device"`
	Inode          string `json:"inode"`
	SecuritySHA256 string `json:"securitySHA256"`
}
type receipt struct {
	Version                 int           `json:"version"`
	Transaction             string        `json:"transaction"`
	SessionKey              string        `json:"sessionKey"`
	PendingSHA256           string        `json:"pendingSHA256"`
	RootDevice              string        `json:"rootDevice"`
	RootInode               string        `json:"rootInode"`
	HomeDevice              string        `json:"homeDevice"`
	HomeInode               string        `json:"homeInode"`
	NamespaceDevice         string        `json:"namespaceDevice"`
	NamespaceInode          string        `json:"namespaceInode"`
	NamespaceSecuritySHA256 string        `json:"namespaceSecuritySHA256"`
	Roles                   []receiptRole `json:"roles"`
}

func hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func hex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func sessionKey(plan, account, root string) string {
	h := sha256.New()
	h.Write([]byte("onboard-kit/session-key/v1\x00"))
	var n [8]byte
	for _, s := range []string{plan, account, root} {
		binary.BigEndian.PutUint64(n[:], uint64(len(s)))
		h.Write(n[:])
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}
func encode(value any) []byte {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(value) != nil {
		return nil
	}
	return b.Bytes()
}
func securityHash(s security) string { return hash(bytes.TrimSuffix(encode(s), []byte("\n"))) }
func decimal(s string) bool {
	if s == "" || len(s) > 1 && s[0] == '0' {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}
func base64Bytes(s string, maximum int) ([]byte, error) {
	if len(s) > ((maximum+2)/3)*4 {
		return nil, errLimit
	}
	data, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || base64.StdEncoding.EncodeToString(data) != s {
		return nil, errContext
	}
	if len(data) > maximum {
		return nil, errLimit
	}
	return data, nil
}
func validUnicodeJSON(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for i := 0; i < len(data); i++ {
		if data[i] != '"' {
			continue
		}
		i++
		for ; i < len(data) && data[i] != '"'; i++ {
			if data[i] != '\\' {
				continue
			}
			i++
			if i >= len(data) {
				return false
			}
			if data[i] != 'u' {
				continue
			}
			if i+4 >= len(data) {
				return false
			}
			v, err := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
			if err != nil {
				return false
			}
			i += 4
			if v >= 0xdc00 && v <= 0xdfff {
				return false
			}
			if v >= 0xd800 && v <= 0xdbff {
				if i+6 >= len(data) || data[i+1] != '\\' || data[i+2] != 'u' {
					return false
				}
				low, err := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
				if err != nil || low < 0xdc00 || low > 0xdfff {
					return false
				}
				i += 6
			}
		}
	}
	return true
}
func jsonValue(d *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, errContext
	}
	token, err := d.Token()
	if err != nil {
		return nil, errContext
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delimiter {
	case '{':
		result := map[string]any{}
		for d.More() {
			k, err := d.Token()
			if err != nil {
				return nil, errContext
			}
			key, ok := k.(string)
			if !ok {
				return nil, errContext
			}
			if _, exists := result[key]; exists {
				return nil, errContext
			}
			value, err := jsonValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return nil, errContext
		}
		return result, nil
	case '[':
		var result []any
		for d.More() {
			value, err := jsonValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		end, err := d.Token()
		if err != nil || end != json.Delim(']') {
			return nil, errContext
		}
		return result, nil
	}
	return nil, errContext
}
func shape(value any, t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		if value == nil {
			return true
		}
		return shape(value, t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := value.(map[string]any)
		if !ok || len(m) != t.NumField() {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			key := f.Tag.Get("json")
			v, ok := m[key]
			if !ok || !shape(v, f.Type) {
				return false
			}
		}
		return true
	case reflect.Slice:
		a, ok := value.([]any)
		if !ok {
			return false
		}
		for _, v := range a {
			if !shape(v, t.Elem()) {
				return false
			}
		}
		return true
	case reflect.String:
		_, ok := value.(string)
		return ok
	case reflect.Bool:
		_, ok := value.(bool)
		return ok
	case reflect.Int:
		n, ok := value.(json.Number)
		return ok && n.String() == "1"
	}
	return false
}
func decode(data []byte, out any, maximum int) error {
	if len(data) > maximum {
		return errLimit
	}
	if !validUnicodeJSON(data) {
		return errContext
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	value, err := jsonValue(d, 0)
	if err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errContext
	}
	if !shape(value, reflect.TypeOf(out).Elem()) {
		return errContext
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(out) != nil {
		return errContext
	}
	return nil
}
func validGuard(g objectGuard, kind string) bool {
	if g.Kind != kind || !decimal(g.Device) || !decimal(g.Inode) || !decimal(g.Security.UID) || !decimal(g.Security.GID) || len(g.Security.Mode) != 4 {
		return false
	}
	for _, c := range g.Security.Mode {
		if c < '0' || c > '7' {
			return false
		}
	}
	s := g.Security
	if s.ACLFormat != "darwin-filesec-v1" && s.ACLFormat != "linux-posix-v1" {
		return false
	}
	a, err := base64Bytes(s.AccessBase64, 65536)
	if err != nil || !s.AccessPresent && len(a) != 0 {
		return false
	}
	a, err = base64Bytes(s.DefaultBase64, 65536)
	if err != nil || !s.DefaultPresent && len(a) != 0 || s.ACLFormat == "darwin-filesec-v1" && (s.DefaultPresent || len(a) != 0) {
		return false
	}
	return true
}
func safeDisplay(s []byte) bool {
	if !utf8.Valid(s) {
		return false
	}
	for _, r := range string(s) {
		if r != '\n' && (!unicode.IsPrint(r) || unicode.Is(unicode.Cf, r) || r == 0x2028 || r == 0x2029) {
			return false
		}
	}
	return true
}
func roleNames() []string { return []string{"local-plan", "preparation", "session-metadata", "record"} }
func validAccount(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || i > 0 && (c == '-' || c == '_' || c == '.')) {
			return false
		}
	}
	return true
}
func (m metadata) valid() bool {
	plan, err := base64Bytes(m.PlanIDBase64, 1048576)
	return err == nil && len(plan) > 0 && utf8.Valid(plan) && validAccount(m.Account) && hex64(m.SessionKey) && sessionKey(string(plan), m.Account, m.Root) == m.SessionKey && safePath(m.Root) && strings.HasPrefix(m.Root, "/") && decimal(m.RootDevice) && decimal(m.RootInode) && relative(m.Record) && relative(m.Preparation) && hex64(m.PlanSHA256) && hex64(m.Transaction)
}
func (p pending) validate() error {
	plan, err := base64Bytes(p.PlanIDBase64, 1048576)
	if err != nil || len(plan) == 0 || !utf8.Valid(plan) || !hex64(p.Transaction) || !validAccount(p.Account) || sessionKey(string(plan), p.Account, p.Root) != p.SessionKey || !safePath(p.Root) || !strings.HasPrefix(p.Root, "/") || !safePath(p.Home) || !strings.HasPrefix(p.Home, "/") || !relative(p.Record) || !relative(p.Preparation) || !validGuard(p.RootGuard, "directory") || !validGuard(p.HomeGuard, "directory") || !private(p.HomeGuard) || len(p.Roles) != 4 {
		return errContext
	}
	display, err := base64Bytes(p.SessionDisplayBase64, displayLimit)
	if err != nil {
		return err
	}
	if !safeDisplay(display) {
		return errContext
	}
	total := 0
	for i, r := range p.Roles {
		if r.Role != roleNames()[i] {
			return errContext
		}
		permitted := false
		switch i {
		case 0, 2:
			permitted = r.Action == "retain" || r.Action == "install" || r.Action == "replace"
		case 1:
			permitted = r.Action == "retain" || r.Action == "install"
		case 3:
			permitted = r.Action == "retain" || r.Action == "create" || r.Action == "reconcile" || r.Action == "bind"
		}
		if !permitted {
			return errContext
		}
		limit := 1048576
		if i == 2 {
			limit = 2097152
		}
		before, err := base64Bytes(r.BeforeBase64, limit)
		if err != nil {
			return err
		}
		after, err := base64Bytes(r.AfterBase64, limit)
		if err != nil {
			return err
		}
		total += len(before) + len(after)
		if total > 10485760 || total+len(display) > 18874368 {
			return errLimit
		}
		if hash(after) != r.AfterSHA256 {
			return errContext
		}
		if r.BeforePresent {
			if r.BeforeGuard == nil || !validGuard(*r.BeforeGuard, "regular") || hash(before) != r.BeforeSHA256 {
				return errContext
			}
		} else if len(before) != 0 || r.BeforeSHA256 != "" || r.BeforeGuard != nil {
			return errContext
		}
		if r.Action == "retain" && (!r.BeforePresent || !bytes.Equal(before, after)) {
			return errContext
		}
		missing := false
		for j, parent := range r.Parents {
			if parent.Ordinal != strconv.Itoa(j) {
				return errContext
			}
			if parent.Present {
				if missing || parent.Guard == nil || !validGuard(*parent.Guard, "directory") {
					return errContext
				}
			} else {
				missing = true
				if parent.Guard != nil {
					return errContext
				}
			}
		}
	}
	return nil
}
