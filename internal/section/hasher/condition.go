package hasher

import "onboard-kit/internal/model"

func hasClauses(condition *model.Condition) bool {
	return condition != nil && len(condition.All) > 0
}

func conditionText(condition model.Condition) string {
	items := make([]string, len(condition.All))
	for index, clause := range condition.All {
		items[index] = clauseText(clause)
	}
	return fields{}.withRaw("all", listText(items), true).text()
}

func clauseText(clause model.Clause) string {
	return fields{}.
		withString("question", clause.Question).
		withRaw("answer", answerTestText(clause.Answer), !isZeroAnswerTest(clause.Answer)).
		text()
}

func isZeroAnswerTest(test model.AnswerTest) bool {
	return !hasValue(test.Is) && !hasValue(test.IsNot) &&
		len(test.IsOneOf) == 0 && len(test.IsNotOneOf) == 0 && test.IsSettled == nil
}

func answerTestText(test model.AnswerTest) string {
	settled := test.IsSettled != nil
	return fields{}.
		withRaw("is", valueText(test.Is), hasValue(test.Is)).
		withRaw("isNot", valueText(test.IsNot), hasValue(test.IsNot)).
		withRaw("isOneOf", valueSetText(test.IsOneOf), len(test.IsOneOf) > 0).
		withRaw("isNotOneOf", valueSetText(test.IsNotOneOf), len(test.IsNotOneOf) > 0).
		withRaw("isSettled", settledText(test.IsSettled), settled).
		text()
}

func settledText(settled *bool) string {
	if settled != nil && *settled {
		return "true"
	}
	return "false"
}
