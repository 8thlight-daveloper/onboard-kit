package model

type ValueKind int

const (
	ValueKindNone ValueKind = iota
	ValueKindText
	ValueKindFlag
	ValueKindList
	ValueKindUnknown
)

type Value struct {
	Kind ValueKind
	Text string
	Flag bool
	List []string
}

type Answers struct {
	Plan     map[string]Value
	Sections map[string]map[string]Value
}
