package resolver

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

func copySection(section model.Section) model.Section {
	copied := section
	copied.Enabled = copyPointer(section.Enabled)
	copied.IncludeIf = copyCondition(section.IncludeIf)
	copied.After = slices.Clone(section.After)
	copied.Escalation = slices.Clone(section.Escalation)
	copied.Variables = copyEach(section.Variables, copyVariable)
	copied.Steps = copyEach(section.Steps, copyStep)
	copied.Source = copyPointer(section.Source)
	copied.UnknownFields = copyUnknown(section.UnknownFields)
	return copied
}

func copyEach[T any](items []T, copyItem func(T) T) []T {
	if items == nil {
		return nil
	}
	copied := make([]T, len(items))
	for index, item := range items {
		copied[index] = copyItem(item)
	}
	return copied
}

func copyPointer[T any](pointer *T) *T {
	if pointer == nil {
		return nil
	}
	copied := *pointer
	return &copied
}

func copyVariable(variable model.Variable) model.Variable {
	copied := variable
	copied.Options = slices.Clone(variable.Options)
	copied.Default = copyValue(variable.Default)
	copied.AskIf = copyCondition(variable.AskIf)
	return copied
}

func copyStep(step model.Step) model.Step {
	copied := step
	copied.Run = slices.Clone(step.Run)
	copied.Env = maps.Clone(step.Env)
	copied.Escalation = slices.Clone(step.Escalation)
	copied.Checks = copyEach(step.Checks, copyCheck)
	return copied
}

func copyCheck(check model.Check) model.Check {
	copied := check
	copied.Run = slices.Clone(check.Run)
	return copied
}

func copyValue(value model.Value) model.Value {
	copied := value
	copied.List = slices.Clone(value.List)
	return copied
}

func copyCondition(condition *model.Condition) *model.Condition {
	if condition == nil {
		return nil
	}
	return &model.Condition{All: copyEach(condition.All, copyClause)}
}

func copyClause(clause model.Clause) model.Clause {
	test := clause.Answer
	test.Is = copyValue(test.Is)
	test.IsNot = copyValue(test.IsNot)
	test.IsOneOf = copyEach(test.IsOneOf, copyValue)
	test.IsNotOneOf = copyEach(test.IsNotOneOf, copyValue)
	test.IsSettled = copyPointer(test.IsSettled)
	clause.Answer = test
	return clause
}

func copyUnknown(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	copied := make(map[string]any, len(fields))
	for key, value := range fields {
		copied[key] = copyAny(value)
	}
	return copied
}

func copyAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return copyUnknown(typed)
	case []any:
		return copyEach(typed, copyAny)
	default:
		return value
	}
}
