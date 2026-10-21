package model

type Date string

type Plan struct {
	SchemaVersion int
	ID            string
	Provenance    Provenance
	Origin        Origin
	URL           string
	ResultsPath   string
	Directory     Directory
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

type Directory struct {
	Contacts []Contact
	Roles    []Role
	Teams    []Team
}

type Contact struct {
	ID      string
	Name    string
	Email   string
	Slack   string
	SlackID string
}

type Role struct {
	ID      string
	Name    string
	Holders []string
}

type Team struct {
	ID      string
	Name    string
	Contact string
	Holders map[string][]string
}
