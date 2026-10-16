package model

// Step is one action or checklist item in a section.
type Step struct {
	ID    string
	Name  string
	Notes string
	Owner string
	// Timing is absent when the step inherits its section's timing.
	Timing    Timing
	Run       []string
	Open      string
	Directory string
	// Env maps a variable name to the text it runs with.
	Env    map[string]string
	Checks []Check
}

// Check says how to tell that a step worked. Exactly one field is set.
type Check struct {
	// Run is a command that must exit with code 0.
	Run []string
	// Ask is a yes or no question.
	Ask string
	// Env is the name of an environment variable that must be set.
	Env string
}
