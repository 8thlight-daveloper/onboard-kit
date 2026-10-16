package model

type ValueKind int

const (
	ValueKindNone ValueKind = iota
	ValueKindText
	ValueKindFlag
	ValueKindCommand
	ValueKindUnknown
)

type Value struct {
	Kind    ValueKind
	Text    string
	Flag    bool
	Command []string
}

type Answers struct {
	Plan     map[string]Value
	Sections map[string]map[string]Value
}
