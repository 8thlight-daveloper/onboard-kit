package joiner

import "onboard-kit/internal/model"

func layerAnswers(base, over model.Answers) model.Answers {
	return model.Answers{
		Plan:     layerScope(base.Plan, over.Plan),
		Sections: layerSections(base.Sections, over.Sections),
	}
}

func layerScope(base, over map[string]model.Value) map[string]model.Value {
	if base == nil && len(over) == 0 {
		return nil
	}
	scope := make(map[string]model.Value, len(base)+len(over))
	for name, value := range base {
		scope[name] = copyValue(value)
	}
	for name, value := range over {
		scope[name] = copyValue(value)
	}
	return scope
}

func layerSections(base, over map[string]map[string]model.Value) map[string]map[string]model.Value {
	if base == nil && !holdsAnswers(over) {
		return nil
	}
	sections := make(map[string]map[string]model.Value, len(base)+len(over))
	for id, scope := range base {
		sections[id] = layerScope(scope, over[id])
	}
	for id, scope := range over {
		if _, known := base[id]; !known && len(scope) > 0 {
			sections[id] = layerScope(nil, scope)
		}
	}
	return sections
}

func holdsAnswers(sections map[string]map[string]model.Value) bool {
	for _, scope := range sections {
		if len(scope) > 0 {
			return true
		}
	}
	return false
}
