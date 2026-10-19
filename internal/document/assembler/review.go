package assembler

import "onboard-kit/internal/model"

func reviewLine(provenance model.Provenance) string {
	return "Generated " + string(provenance.GeneratedOn) + " from template set `" + provenance.TemplateSet +
		"`. A person must review this before it is sent."
}
