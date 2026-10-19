package evaluator

import "onboard-kit/internal/model"

func answerTestHolds(test model.AnswerTest, answer model.Value, settled bool) bool {
	if test.IsSettled != nil && *test.IsSettled != settled {
		return false
	}
	if !hasValueTest(test) {
		return true
	}
	return settled && valueTestsHold(test, answer)
}

func hasValueTest(test model.AnswerTest) bool {
	return test.Is.Kind != model.ValueKindNone ||
		test.IsNot.Kind != model.ValueKindNone ||
		test.IsOneOf != nil ||
		test.IsNotOneOf != nil
}

func valueTestsHold(test model.AnswerTest, answer model.Value) bool {
	if isList(answer) || usesList(test) {
		return false
	}
	if test.Is.Kind != model.ValueKindNone && !isEqual(answer, test.Is) {
		return false
	}
	if test.IsNot.Kind != model.ValueKindNone && isEqual(answer, test.IsNot) {
		return false
	}
	if test.IsOneOf != nil && !isEqualToAny(answer, test.IsOneOf) {
		return false
	}
	if test.IsNotOneOf != nil && isEqualToAny(answer, test.IsNotOneOf) {
		return false
	}
	return true
}

func usesList(test model.AnswerTest) bool {
	return isList(test.Is) || isList(test.IsNot) || containsList(test.IsOneOf) || containsList(test.IsNotOneOf)
}
