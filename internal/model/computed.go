package model

// Question is what the finder returns for one active question. It is computed
// and never stored.
type Question struct {
	// ID is <scope>.<name>, such as access.repo_url.
	ID          string
	Scope       string
	Name        string
	Description string
	Type        VariableType
	Options     []string
	Required    bool
	Phase       Phase
	// Answer is the stored answer. A zero Value means none.
	Answer Value
	// Default is shown to the person and is not a stored answer.
	Default Value
	// Group is the key shared by questions with the same name, type, and
	// options.
	Group       string
	Suggestions []Suggestion
}

// Suggestion is a value offered for a question.
type Suggestion struct {
	Value  Value
	Source SuggestionSource
}

// Blocker is an active, required planning question with no answer.
type Blocker struct {
	Question string
}

// Problem is a report about a plan or one value. It is a normal result, not a
// failure.
type Problem struct {
	// Code is a stable machine name that the reporting component's contract
	// defines.
	Code     string
	Message  string
	Question string
	Section  string
	Detail   map[string]string
}
