package model

// Section is the content unit in a plan. A section template is a Section with
// no Source.
type Section struct {
	ID           string
	Title        string
	Summary      string
	Owner        string
	Timing       Timing
	Verification string
	Body         string
	IncludeIf    *Condition
	After        []string
	Variables    []Variable
	Steps        []Step
	// Source is nil on a template and on a section added by hand.
	Source *Source
	// Extra holds fields the tool does not know. A plan keeps them.
	Extra map[string]any
}

// Source records the template a section was copied from.
type Source struct {
	Template string
	// Hash is the content hash of the section, computed over everything
	// except Source.
	Hash string
}
