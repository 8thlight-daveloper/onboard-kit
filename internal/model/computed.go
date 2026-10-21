package model

type Question struct {
	ID              string
	Scope           string
	Name            string
	Description     string
	Example         string
	Type            VariableType
	Options         []string
	Choices         []Choice
	Required        bool
	Multiple        bool
	Audience        Audience
	Answer          Value
	DeclaredDefault Value
	Group           string
	Suggestions     []Suggestion
}

type Choice struct {
	Value string
	Label string
}

type Suggestion struct {
	Value  Value
	Source SuggestionSource
}

type Blocker struct {
	Question string
}

type Problem struct {
	Code     string
	Severity Severity
	Message  string
	Question string
	Section  string
	Detail   map[string]string
}
