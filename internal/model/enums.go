package model

// VariableType is the kind of answer a variable takes. Absent means string.
type VariableType string

// The variable types.
const (
	VariableTypeString  VariableType = "string"
	VariableTypeBoolean VariableType = "boolean"
	VariableTypeChoice  VariableType = "choice"
	VariableTypeCommand VariableType = "command"
)

// Phase says which process asks a variable. Absent means planning.
type Phase string

// The phases.
const (
	PhasePlanning   Phase = "planning"
	PhaseOnboarding Phase = "onboarding"
)

// Timing says when a section or step happens. Absent means anytime.
type Timing string

// The timings, in the order sections sort.
const (
	TimingBeforeDayOne Timing = "before-day-one"
	TimingWeekOne      Timing = "week-one"
	TimingFirst30Days  Timing = "first-30-days"
	TimingDays30To90   Timing = "days-30-to-90"
	TimingAnytime      Timing = "anytime"
)

// SuggestionSource says where a suggested value came from.
type SuggestionSource string

// The suggestion sources.
const (
	SuggestionSourceDefault SuggestionSource = "default"
	SuggestionSourceGiven   SuggestionSource = "given"
)
