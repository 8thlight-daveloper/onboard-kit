package assembler

import "onboard-kit/internal/model"

type HelpReference struct {
	Kind string
	ID   string
}

type Context struct {
	Account     string
	GeneralHelp []HelpReference
}

type Assembler interface {
	Assemble(Context, model.Plan, model.Preparation) (string, []model.Problem)
}
