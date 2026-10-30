package editguide

import (
	"bytes"
	"fmt"
	"sort"

	editorcontract "onboard-kit/contracts/catalog/editor"
	materializercontract "onboard-kit/contracts/catalog/materializer"
)

const (
	unavailable      = "Review unavailable.\n"
	unavailableLimit = "Review unavailable: preserve all plan data within supported limits.\n"
	unavailableGuide = "Review unavailable: preparation guidance sources conflict.\n"
)

type frozenSave struct {
	save      []byte
	unchanged bool
}

type guide struct {
	markdown []byte
	reasons  map[string]bool
}

func (s *session) refuse(text string, problems []problem) (frozenSave, bool) {
	if problems != nil {
		text += problemText(problems)
	}
	s.write(text)
	return frozenSave{}, false
}

func (s *session) review() (frozenSave, bool) {
	proposal, diagnostics := s.command.editor.Edit(s.original, s.operations)
	warnings := editorProblems(diagnostics)
	if hasErrors(warnings) {
		return s.refuse(unavailable, warnings)
	}
	authoring := s.command.authoring.Validate(proposal)
	if hasModelErrors(authoring) {
		return s.refuse(unavailable, modelProblems("authoring", authoring))
	}
	checked := s.command.checker.Validate(proposal)
	if hasModelErrors(checked) {
		return s.refuse(unavailable, modelProblems("plan", checked))
	}
	warnings = append(append(warnings, modelProblems("authoring", authoring)...), modelProblems("plan", checked)...)
	encoded, problems := s.command.encoder.Encode(proposal)
	if hasModelErrors(problems) {
		return s.refuse(unavailable, modelProblems("encode", problems))
	}
	if len(encoded) > maxSourceBytes {
		return s.refuse(unavailableLimit, nil)
	}
	decoded, problems := s.command.decoder.Decode(encoded)
	if hasModelErrors(problems) {
		return s.refuse(unavailable, modelProblems("decode", problems))
	}
	if !equivalent(proposal, decoded) {
		return s.refuse(unavailableLimit, nil)
	}
	guides, guidanceWarnings, failure := s.guidance()
	if failure != "" {
		return s.refuse(failure, guidanceWarnings)
	}
	frozen := frozenSave{save: encoded}
	if equivalent(proposal, s.original) {
		frozen = frozenSave{save: s.selection.raw, unchanged: true}
	}
	var d display
	d.add("Catalog edit review — not saved. No personal work or outcomes are changed.\nPlan: ")
	d.scalar(s.selection.path)
	d.add("\n")
	d.operations(s.operations)
	d.document("Original plan bytes", s.selection.raw)
	d.document("Proposed plan bytes", frozen.save)
	if frozen.unchanged {
		d.add("Save action: unchanged; original bytes retained.\n")
	} else {
		d.add("Save action: replace.\n")
	}
	d.problems(append(warnings, guidanceWarnings...))
	d.guides(guides)
	d.add("Advice includes retired new additions; no arrangement or activity is completed.\n" + reviewPrompt)
	if d.full {
		return s.refuse(unavailableLimit, nil)
	}
	if !s.write(d.text.String()) {
		return frozenSave{}, false
	}
	return frozen, true
}

func (s *session) guidance() (map[string]*guide, []problem, string) {
	guides := map[string]*guide{}
	var warnings []problem
	for _, operation := range s.operations {
		if operation.Kind != editorcontract.Add {
			continue
		}
		bindings := make([]materializercontract.Binding, len(operation.Bindings))
		for i, binding := range operation.Bindings {
			bindings[i] = materializercontract.Binding{Name: binding.Name, Value: binding.Value}
		}
		result, diagnostics := s.command.materializer.Create(materializercontract.Request{
			PlanID:    s.original.ID,
			Directory: s.original.Directory,
			Instances: []materializercontract.Instance{{SourceID: operation.SourceID, PermanentSectionID: operation.SectionID, Bindings: bindings}},
		})
		projected := guidanceProblems(diagnostics)
		if hasErrors(projected) {
			return nil, projected, unavailable
		}
		warnings = append(warnings, projected...)
		for _, advice := range result.Guidance {
			held, found := guides[advice.ID]
			if found && !bytes.Equal(held.markdown, advice.Markdown) {
				return nil, nil, unavailableGuide
			}
			if !found {
				held = &guide{markdown: advice.Markdown, reasons: map[string]bool{}}
				guides[advice.ID] = held
			}
			for _, id := range advice.InstanceIDs {
				held.reasons[id] = true
			}
		}
	}
	return guides, warnings, ""
}

func (d *display) guides(guides map[string]*guide) {
	ids := make([]string, 0, len(guides))
	for id := range guides {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	d.add(fmt.Sprintf("Preparation guidance for newly added instances only: %d\n", len(ids)))
	for _, id := range ids {
		reasons := make([]string, 0, len(guides[id].reasons))
		for reason := range guides[id].reasons {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		d.add("Guide: ")
		d.scalar(id)
		d.add(fmt.Sprintf("\nInstance reasons: %d\n", len(reasons)))
		for _, reason := range reasons {
			d.add("  - ")
			d.scalar(reason)
			d.add("\n")
		}
		d.document("Advisory instructions", guides[id].markdown)
	}
}
