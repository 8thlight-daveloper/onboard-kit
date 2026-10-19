package runner

import "fmt"

type Failure struct {
	Step        int
	Expectation string
	Message     string
}

func (failure Failure) String() string {
	return fmt.Sprintf("step %d, %s: %s", failure.Step, failure.Expectation, failure.Message)
}
