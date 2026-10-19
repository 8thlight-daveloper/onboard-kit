package model

type Condition struct {
	All []Clause
}

type Clause struct {
	Question string
	Answer   AnswerTest
}

type AnswerTest struct {
	Is         Value
	IsNot      Value
	IsOneOf    []Value
	IsNotOneOf []Value
	IsSettled  *bool
}
