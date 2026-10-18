package hasher_test

import (
	contract "onboard-kit/contracts/section/hasher"
	implementation "onboard-kit/internal/section/hasher"
)

func newHasher() contract.Hasher {
	return implementation.New()
}
