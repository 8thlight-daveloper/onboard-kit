package store

import (
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
)

type loadedState struct {
	selection          selection
	account            string
	destination        string
	preparation        fileSnapshot
	directoriesMissing bool
}

func (l *loadedState) Account() string { return l.account }

func (l *loadedState) Destination() string { return l.destination }

func (l *loadedState) Plan() model.Plan {
	value, _ := plandecoder.New().Decode(l.selection.plan.content)
	return value
}

func (l *loadedState) Preparation() model.Preparation {
	if !l.preparation.exists {
		return model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: l.account, PlanID: l.selection.planValue.ID}
	}
	value, _ := preparationdecoder.New().Decode(l.preparation.content)
	return value
}

func (l *loadedState) Exists() bool { return l.preparation.exists }

func (l *loadedState) Bytes() []byte { return append([]byte(nil), l.preparation.content...) }

func (l *loadedState) DirectoriesMissing() bool { return l.directoriesMissing }
