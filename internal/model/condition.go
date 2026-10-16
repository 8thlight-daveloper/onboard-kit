package model

// Condition holds when a question is asked or a section is included. Every
// clause must hold.
type Condition struct {
	All []Clause
}

// Clause tests the answer to one question.
type Clause struct {
	// Question is a question id. A bare name means the same section.
	Question string
	Answer   AnswerTest
}

// AnswerTest holds the tests on one answer. All the tests that are set must
// hold. A zero Value means the test is not set.
type AnswerTest struct {
	Is         Value
	IsNot      Value
	IsOneOf    []Value
	IsNotOneOf []Value
	// IsAnswered is nil when the test is not set.
	IsAnswered *bool
}
