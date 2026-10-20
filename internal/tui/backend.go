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
