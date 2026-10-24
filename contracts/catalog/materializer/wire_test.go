package materializer_test

import (
	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/catalog/materializer"
)

func newMaterializer() api.Materializer { return materializer.New() }
