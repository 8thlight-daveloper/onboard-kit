package encoder

import "onboard-kit/internal/model"

func originNode(origin model.Origin) node {
	builder := &mappingBuilder{}
	builder.add("planId", stringField(origin.PlanID))
	builder.add("source", stringField(origin.Source))
	builder.add("contentHash", stringField(origin.ContentHash))
	builder.add("refreshedOn", stringField(string(origin.RefreshedOn)))
	return builder.build()
}
