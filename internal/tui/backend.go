package tui

import (
	"onboard-kit/internal/document/assembler"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/editor"
	"onboard-kit/internal/plan/validator"
	"onboard-kit/internal/question/finder"
)

type Backend interface {
	Questions(plan model.Plan, audience model.Audience) []model.Question
	Blockers(plan model.Plan, audience model.Audience) []model.Blocker
	Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem)
	SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem)
	SetContact(plan model.Plan, contact model.Contact) (model.Plan, []model.Problem)
	SetRole(plan model.Plan, role model.Role) (model.Plan, []model.Problem)
	SetTeam(plan model.Plan, team model.Team) (model.Plan, []model.Problem)
	RemoveContact(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem)
	RemoveRole(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem)
	RemoveTeam(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem)
	Validate(plan model.Plan) []model.Problem
	Assemble(plan model.Plan) (string, []model.Problem)
}

type components struct {
	finder    finder.Finder
	editor    editor.Editor
	validator validator.Validator
	assembler assembler.Assembler
}

func NewBackend() Backend {
	return components{finder: finder.New(), editor: editor.New(), validator: validator.New(), assembler: assembler.New()}
}

func (parts components) Validate(plan model.Plan) []model.Problem {
	return parts.validator.Validate(plan)
}

func (parts components) Questions(plan model.Plan, audience model.Audience) []model.Question {
	return parts.finder.Questions(plan, audience)
}

func (parts components) Blockers(plan model.Plan, audience model.Audience) []model.Blocker {
	return parts.finder.Blockers(plan, audience)
}

func (parts components) Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem) {
	return parts.editor.Submit(plan, question, value)
}

func (parts components) SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem) {
	return parts.editor.SetSectionEnabled(plan, section, enabled)
}

func (parts components) Assemble(plan model.Plan) (string, []model.Problem) {
	return parts.assembler.Assemble(plan)
}

func (parts components) SetContact(plan model.Plan, contact model.Contact) (model.Plan, []model.Problem) {
	return parts.editor.SetContact(plan, contact)
}

func (parts components) SetRole(plan model.Plan, role model.Role) (model.Plan, []model.Problem) {
	return parts.editor.SetRole(plan, role)
}

func (parts components) SetTeam(plan model.Plan, team model.Team) (model.Plan, []model.Problem) {
	return parts.editor.SetTeam(plan, team)
}

func (parts components) RemoveContact(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return parts.editor.RemoveContact(plan, id)
}

func (parts components) RemoveRole(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return parts.editor.RemoveRole(plan, id)
}

func (parts components) RemoveTeam(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	return parts.editor.RemoveTeam(plan, id)
}
