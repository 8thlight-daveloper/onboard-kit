package model

type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeBoolean VariableType = "boolean"
	VariableTypeChoice  VariableType = "choice"
	VariableTypeCommand VariableType = "command"
	VariableTypeContact VariableType = "contact"
	VariableTypeRole    VariableType = "role"
	VariableTypeTeam    VariableType = "team"

	DefaultVariableType = VariableTypeString
)

type Audience string

const (
	AudienceOnboarder Audience = "onboarder"
	AudienceOnboardee Audience = "onboardee"

	DefaultAudience = AudienceOnboarder
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
