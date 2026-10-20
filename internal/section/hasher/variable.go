package hasher

import "onboard-kit/internal/model"

func variablesText(variables []model.Variable) string {
	items := make([]string, len(variables))
	for index, variable := range variables {
		items[index] = variableText(variable)
	}
	return listText(items)
}

func variableText(variable model.Variable) string {
	variableType := variable.Type
	if variableType == "" {
		variableType = model.DefaultVariableType
	}
	audience := variable.Audience
	if audience == "" {
		audience = model.DefaultAudience
	}
	asked := hasClauses(variable.AskIf)
	return fields{}.
		withString("name", variable.Name).
		withProse("description", variable.Description).
		withProse("example", variable.Example).
		withString("type", string(variableType)).
		withFlag("multiple", variable.Multiple).
		withStrings("options", variable.Options).
		withString("pattern", variable.Pattern).
		withProse("patternHint", variable.PatternHint).
		withFlag("required", variable.Required).
		withRaw("default", valueText(variable.Default), hasValue(variable.Default)).
		withString("audience", string(audience)).
		withRaw("askIf", conditionTextOrEmpty(variable.AskIf), asked).
		text()
}

func conditionTextOrEmpty(condition *model.Condition) string {
	if condition == nil {
		return ""
	}
	return conditionText(*condition)
}
