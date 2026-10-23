package scheduler_test

import (
	contract "onboard-kit/contracts/section/scheduler"
	implementation "onboard-kit/internal/section/scheduler"
)

func newScheduler() contract.Scheduler {
	return implementation.New()
}
