package editor_test

import (
	contract "onboard-kit/contracts/preparation/editor"
	implementation "onboard-kit/internal/preparation/editor"
)

func newEditor() contract.Editor { return implementation.New() }
