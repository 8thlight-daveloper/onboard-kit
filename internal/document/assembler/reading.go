package assembler

import (
	"slices"

	"onboard-kit/internal/model"
)

func (a Assembler) readsVariable(section model.Section, name string, context planContext) bool {
	probe := withUnansweredVariable(section, name)
	answers := model.Answers{Sections: map[string]map[string]model.Value{
		section.ID: answersWithout(context.answers.Sections[section.ID], name),
	}}
	question := section.ID + "." + name
	for _, candidate := range readingSections(probe) {
		_, problems := a.resolver.Resolve(candidate, answers, context.directory, context.team)
		if slices.ContainsFunc(problems, func(problem model.Problem) bool { return problem.Question == question }) {
			return true
		}
	}
	return false
}

func readingSections(section model.Section) []model.Section {
	sections := []model.Section{section}
	for _, item := range section.Escalation {
		sections = append(sections, itemSection(section, item))
	}
	for _, step := range section.Steps {
		for _, item := range step.Escalation {
			sections = append(sections, itemSection(section, item))
		}
	}
	return sections
}

func withUnansweredVariable(section model.Section, name string) model.Section {
	variables := slices.Clone(section.Variables)
	for index := range variables {
		if variables[index].Name == name {
			variables[index].Audience = ""
			variables[index].Default = model.Value{}
		}
	}
	section.Variables = variables
	return section
}

func answersWithout(answers map[string]model.Value, name string) map[string]model.Value {
	kept := make(map[string]model.Value, len(answers))
	for key, value := range answers {
		if key != name {
			kept[key] = value
		}
	}
	return kept
}
