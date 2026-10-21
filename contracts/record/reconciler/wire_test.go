package reconciler_test

import (
	contract "onboard-kit/contracts/record/reconciler"
	implementation "onboard-kit/internal/record/reconciler"
)

func newReconciler() contract.Reconciler {
	return implementation.New()
}
