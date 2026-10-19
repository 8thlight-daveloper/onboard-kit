package instantiator

import "onboard-kit/internal/model"

func copyVariables(variables []model.Variable) []model.Variable {
	return copySlice(variables, copyVariable)
}

func copyVariable(variable model.Variable) model.Variable {
	variable.Options = copyStrings(variable.Options)
	variable.Default = copyValue(variable.Default)
	variable.AskIf = copyCondition(variable.AskIf)
	return variable
}

func copyValue(value model.Value) model.Value {
	value.List = copyStrings(value.List)
	return value
}

func copyCondition(condition *model.Condition) *model.Condition {
	if condition == nil {
		return nil
	}
	return &model.Condition{All: copySlice(condition.All, copyClause)}
}

func copyClause(clause model.Clause) model.Clause {
	test := clause.Answer
	test.Is = copyValue(test.Is)
	test.IsNot = copyValue(test.IsNot)
	test.IsOneOf = copySlice(test.IsOneOf, copyValue)
	test.IsNotOneOf = copySlice(test.IsNotOneOf, copyValue)
	if test.IsSettled != nil {
		settled := *test.IsSettled
		test.IsSettled = &settled
	}
	clause.Answer = test
	return clause
}
