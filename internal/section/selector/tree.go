package selector

import "onboard-kit/internal/model"

func arrangeTree(active []model.Section) []model.Section {
	var roots []model.Section
	childrenOf := map[string][]model.Section{}
	for _, section := range active {
		if section.Parent == "" {
			roots = append(roots, section)
			continue
		}
		childrenOf[section.Parent] = append(childrenOf[section.Parent], section)
	}
	arranged := make([]model.Section, 0, len(active))
	return appendSubtrees(arranged, orderSiblings(roots, model.DefaultTiming), model.DefaultTiming, childrenOf)
}

func appendSubtrees(arranged []model.Section, siblings []model.Section, inherited model.Timing, childrenOf map[string][]model.Section) []model.Section {
	for _, section := range siblings {
		arranged = append(arranged, section)
		timing := effectiveTiming(section, inherited)
		children := orderSiblings(childrenOf[section.ID], timing)
		arranged = appendSubtrees(arranged, children, timing, childrenOf)
	}
	return arranged
}
