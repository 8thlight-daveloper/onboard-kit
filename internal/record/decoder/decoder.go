package decoder

import "onboard-kit/internal/model"

type Decoder struct{}

func New() Decoder {
	return Decoder{}
}

func (Decoder) Decode(data []byte) (model.Record, []model.Problem) {
	root, fileCode := parseDocument(data)
	if fileCode != "" {
		return model.Record{}, []model.Problem{fileProblem(fileCode)}
	}
	if versionProblems := checkVersion(root); versionProblems != nil {
		return model.Record{}, versionProblems
	}
	if kindProblems := checkKind(root); kindProblems != nil {
		return model.Record{}, kindProblems
	}
	reading := &collector{}
	record := reading.recordFrom(root)
	if len(reading.findings) > 0 {
		return model.Record{}, reading.sortedProblems()
	}
	return record, nil
}
