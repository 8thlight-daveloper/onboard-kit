package model

type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeBoolean VariableType = "boolean"
	VariableTypeChoice  VariableType = "choice"
	VariableTypeCommand VariableType = "command"
	VariableTypePerson  VariableType = "person"

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
	TimingDays8To30    Timing = "days-8-to-30"
	TimingDays31To90   Timing = "days-31-to-90"
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
