package validator

import "onboard-kit/internal/model"

type Context struct {
	PlanID   string
	Account  string
	Contacts []string
	Teams    []string
	Sections []string
	Record   *RecordIdentity
}

type RecordIdentity struct {
	Account string
	PlanID  string
}

type Validator interface {
	Validate(Context, model.Preparation) []model.Problem
}
