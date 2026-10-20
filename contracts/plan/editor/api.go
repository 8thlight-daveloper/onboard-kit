package editor

import "onboard-kit/internal/model"

type Editor interface {
	Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem)
	SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem)
}
