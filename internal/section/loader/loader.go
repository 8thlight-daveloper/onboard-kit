package loader

import "onboard-kit/internal/model"

type Loader struct{}

func New() Loader {
	return Loader{}
}

func (Loader) Load(data []byte) (model.Section, []model.Problem) {
	root, fileCode := parseDocument(data)
	if fileCode != "" {
		return model.Section{}, []model.Problem{fileProblem(fileCode)}
	}
	if versionProblems := checkVersion(root); versionProblems != nil {
		return model.Section{}, versionProblems
	}
	reading := &collector{}
	section := reading.sectionFrom(root)
	if len(reading.findings) > 0 {
		return model.Section{}, reading.sortedProblems()
	}
	return section, nil
}
