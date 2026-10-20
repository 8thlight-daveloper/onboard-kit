package model

type Date string

type Plan struct {
	SchemaVersion int
	ID            string
	Provenance    Provenance
	Origin        Origin
	Variables     []Variable
	Answers       Answers
	Sections      []Section
	UnknownFields map[string]any
}

type Provenance struct {
	TemplateSet string
	GeneratedOn Date
}

type Origin struct {
	PlanID      string
	Source      string
	ContentHash string
	RefreshedOn Date
}
