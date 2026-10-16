package model

type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeBoolean VariableType = "boolean"
	VariableTypeChoice  VariableType = "choice"
	VariableTypeCommand VariableType = "command"
	VariableTypePeople  VariableType = "people"

	DefaultVariableType = VariableTypeString
)

type Phase string

const (
	PhasePlanning   Phase = "planning"
	PhaseOnboarding Phase = "onboarding"

	DefaultPhase = PhasePlanning
)

type Timing string

const (
	TimingBeforeDayOne Timing = "before-day-one"
	TimingWeekOne      Timing = "week-one"
	TimingFirst30Days  Timing = "first-30-days"
	TimingDays30To90   Timing = "days-30-to-90"
	TimingAnytime      Timing = "anytime"

	DefaultTiming = TimingAnytime
)

type SuggestionSource string

const (
	SuggestionSourceDefault SuggestionSource = "default"
	SuggestionSourceGiven   SuggestionSource = "given"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)
