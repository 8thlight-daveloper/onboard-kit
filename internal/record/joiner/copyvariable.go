package joiner

import "onboard-kit/internal/model"

func copyVariable(variable model.Variable) model.Variable {
	variable.Options = copyEach(variable.Options, identity[string])
	variable.Default = copyValue(variable.Default)
	variable.AskIf = copyCondition(variable.AskIf)
	return variable
}

func copyValue(value model.Value) model.Value {
	value.List = copyEach(value.List, identity[string])
	return value
}

func copyCondition(condition *model.Condition) *model.Condition {
	if condition == nil {
		return nil
	}
	return &model.Condition{All: copyEach(condition.All, copyClause)}
}

func copyClause(clause model.Clause) model.Clause {
	clause.Answer = copyAnswerTest(clause.Answer)
	return clause
}

func copyAnswerTest(test model.AnswerTest) model.AnswerTest {
	test.Is = copyValue(test.Is)
	test.IsNot = copyValue(test.IsNot)
	test.IsOneOf = copyEach(test.IsOneOf, copyValue)
	test.IsNotOneOf = copyEach(test.IsNotOneOf, copyValue)
	test.IsSettled = copyPointer(test.IsSettled)
	return test
}
