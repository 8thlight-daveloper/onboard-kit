package editor_test

import (
	contract "onboard-kit/contracts/plan/editor"
	implementation "onboard-kit/internal/plan/editor"
)

func newEditor() contract.Editor {
	return implementation.New()
}
