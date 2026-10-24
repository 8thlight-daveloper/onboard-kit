package editor_test

import (
	contract "onboard-kit/contracts/catalog/editor"
	implementation "onboard-kit/internal/catalog/editor"
)

func newEditor() contract.Editor { return implementation.New() }
