package model

type Preparation struct {
	SchemaVersion int
	Kind          string
	Account       string
	PlanID        string
	Recipient     PreparationRecipient
	Arrangements  []Arrangement
	UnknownFields map[string]any
}

type PreparationRecipient struct {
	Name             string
	CompanyEmail     string
	PlannedStartDate Date
	Team             string
	Manager          string
	OnboardingOwner  string
	Buddy            string
	WorkLocation     string
	TimeZone         string
}

type Arrangement struct {
	ID         string
	Section    string
	Kind       string
	Owner      string
	Contacts   []string
	Reference  string
	TargetDate Date
	Status     string
	Note       string
}
