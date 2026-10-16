package model

// Plan is the self-contained description of one organization's onboarding.
// A Plan is a value. No function changes one, and an edit returns a new Plan.
type Plan struct {
	SchemaVersion int
	Provenance    Provenance
	// Variables are the plan variables: the axis variables, document details,
	// and settings.
	Variables []Variable
	Answers   Answers
	Sections  []Section
	// Extra holds fields the tool does not know. They are kept on load and
	// written back.
	Extra map[string]any
}

// Provenance records where a plan came from.
type Provenance struct {
	// TemplateSet is the build's commit sha or tag.
	TemplateSet string
	// GeneratedOn is an ISO date, such as 2026-10-15. A driver passes it in,
	// because the core never reads a clock.
	GeneratedOn string
}
