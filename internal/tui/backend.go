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

func (c components) Validate(plan model.Plan) []model.Problem { return c.validator.Validate(plan) }

func (c components) Questions(plan model.Plan, audience model.Audience) []model.Question {
	return c.finder.Questions(plan, audience)
}

func (c components) Blockers(plan model.Plan, audience model.Audience) []model.Blocker {
	return c.finder.Blockers(plan, audience)
}

func (c components) Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem) {
	return c.editor.Submit(plan, question, value)
}

func (c components) SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem) {
	return c.editor.SetSectionEnabled(plan, section, enabled)
}

func (c components) Assemble(plan model.Plan) (string, []model.Problem) {
	return c.assembler.Assemble(plan)
}
