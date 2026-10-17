package decoder

import "onboard-kit/internal/model"

type Decoder struct{}

func New() Decoder {
	return Decoder{}
}

func (Decoder) Decode(data []byte) (model.Plan, []model.Problem) {
	root, fileCode := parseDocument(data)
	if fileCode != "" {
		return model.Plan{}, []model.Problem{fileProblem(fileCode)}
	}
	if versionProblems := checkVersion(root); versionProblems != nil {
		return model.Plan{}, versionProblems
	}
	reading := &collector{}
	plan := reading.planFrom(root)
	if len(reading.findings) > 0 {
		return model.Plan{}, reading.sortedProblems()
	}
	return plan, nil
}
