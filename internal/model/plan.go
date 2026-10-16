package model

type Date string

type Plan struct {
	SchemaVersion int
	Provenance    Provenance
	Variables     []Variable
	Answers       Answers
	Sections      []Section
	UnknownFields map[string]any
}

type Provenance struct {
	TemplateSet string
	GeneratedOn Date
}
