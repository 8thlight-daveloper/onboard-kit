package editor

import "onboard-kit/internal/model"

type Cleared struct {
	Answers []string
	Holders []string
}

type Editor interface {
	Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem)
	SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem)
	SetContact(plan model.Plan, contact model.Contact) (model.Plan, []model.Problem)
	SetRole(plan model.Plan, role model.Role) (model.Plan, []model.Problem)
	SetTeam(plan model.Plan, team model.Team) (model.Plan, []model.Problem)
	RemoveContact(plan model.Plan, id string) (model.Plan, Cleared, []model.Problem)
	RemoveRole(plan model.Plan, id string) (model.Plan, Cleared, []model.Problem)
	RemoveTeam(plan model.Plan, id string) (model.Plan, Cleared, []model.Problem)
}
