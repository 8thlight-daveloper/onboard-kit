package model

// ValueKind says which variant a Value holds.
type ValueKind int

// The variants of a Value. The zero kind means absent.
const (
	ValueKindNone ValueKind = iota
	ValueKindText
	ValueKindFlag
	ValueKindCommand
	ValueKindUnknown
)

// Value is one piece of data: text, a flag, a command, or the unknown marker
// that the file writes as unknown-yet. Only the field for Kind is meaningful.
type Value struct {
	Kind    ValueKind
	Text    string
	Flag    bool
	Command []string
}

// Answers holds the stored answers. Plan maps a plan variable name to its
// answer. Sections maps a section id to that section's own answers.
type Answers struct {
	Plan     map[string]Value
	Sections map[string]map[string]Value
}
