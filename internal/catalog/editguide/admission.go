package editguide

import (
	"onboard-kit/internal/model"
)

func (c command) admit(raw []byte) (model.Plan, string, string) {
	original, problems := c.decoder.Decode(raw)
	if hasModelErrors(problems) {
		return model.Plan{}, problemText(modelProblems("decode", problems)), "data"
	}
	proposal, diagnostics := c.editor.Edit(original, nil)
	if projected := editorProblems(diagnostics); hasErrors(projected) {
		return model.Plan{}, problemText(projected), "data"
	}
	encoded, problems := c.encoder.Encode(proposal)
	if hasModelErrors(problems) {
		return model.Plan{}, problemText(modelProblems("encode", problems)), "fidelity"
	}
	if len(encoded) > maxSourceBytes {
		return model.Plan{}, "", "fidelity"
	}
	decoded, problems := c.decoder.Decode(encoded)
	if hasModelErrors(problems) {
		return model.Plan{}, problemText(modelProblems("decode", problems)), "fidelity"
	}
	if !equivalent(proposal, decoded) {
		return model.Plan{}, "", "fidelity"
	}
	if problems := c.authoring.Validate(proposal); hasModelErrors(problems) {
		return model.Plan{}, problemText(modelProblems("authoring", problems)), "data"
	}
	if problems := c.checker.Validate(proposal); hasModelErrors(problems) {
		return model.Plan{}, problemText(modelProblems("plan", problems)), "data"
	}
	return original, "", ""
}
