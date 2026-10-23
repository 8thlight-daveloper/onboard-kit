package recorder_test

import (
	contract "onboard-kit/contracts/outcome/recorder"
	implementation "onboard-kit/internal/outcome/recorder"
)

func newRecorder() contract.Recorder {
	return implementation.New()
}
