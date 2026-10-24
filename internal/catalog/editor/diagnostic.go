package editor

import (
	contract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
	"strings"
)

func refuse(diagnostic contract.Diagnostic) (model.Plan, []contract.Diagnostic) {
	return model.Plan{}, []contract.Diagnostic{diagnostic}
}

func admissionPath(code string) string {
	if code == "edit.held-shape" {
		return "plan"
	}
	return "request"
}

func own(code, stage, path string) contract.Diagnostic {
	message := ""
	switch code {
	case "edit.limit":
		message = "Reduce held plan or edit data to the supported limits."
	case "edit.held-shape":
		message = "Supply ordinary decoded metadata values without unsupported objects."
	case "edit.text":
		message = "Use valid UTF8 held plan and edit data."
	case "edit.plan-id":
		message = "Select an existing plan with an explicit identity."
	case "edit.operation":
		message = "Use exactly the fields of one supported explicit edit operation."
	case "edit.value-kind":
		message = "Choose a supported submission value kind."
	case "edit.reserved-id":
		message = "Choose a new permanent ID; held definitions and answer scopes reserve their IDs."
	case "edit.section":
		message = "Select an exact section already held at this operation."
	case "edit.binding":
		message = "Select an exact local variable declared in this section."
	case "edit.owner":
		message = "Change only onboarder shared facts; preserve newcomer-owned answers."
	}
	return contract.Diagnostic{Stage: stage, Path: path, Code: code, Severity: model.SeverityError, Message: message}
}

func hasErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func diagnosticErrors(diagnostics []contract.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func project(problems []model.Problem, stage, path string) []contract.Diagnostic {
	var diagnostics []contract.Diagnostic
	for _, problem := range problems {
		location := path
		if (stage == "input-authoring" || stage == "final-authoring") && problem.Detail["path"] != "" {
			location += "." + problem.Detail["path"]
		}
		diagnostics = append(diagnostics, diagnostic(problem, stage, location))
	}
	return diagnostics
}

func diagnostic(problem model.Problem, stage, path string) contract.Diagnostic {
	return contract.Diagnostic{Stage: stage, Path: path, Code: problem.Code, Severity: problem.Severity, Message: repairMessage(problem.Code, stage)}
}

func repairMessage(code, stage string) string {
	switch code {
	case "catalog.unknown-id":
		return "Choose an exact section-template ID from catalog list."
	case "materialize.carrier":
		return "Select a section-template source for an activity."
	case "materialize.instance-id":
		return "Supply a nonreserved lowercase permanent section ID."
	case "materialize.duplicate-binding":
		return "Supply each scoped binding name once."
	case "materialize.binding-name":
		return "Bind an exact variable declared in this scope."
	case "materialize.binding-owner":
		return "Keep newcomer answers outside the shared organization plan."
	case "answer.inactive-question":
		return "Reactivate or satisfy this question's declared applicability before submitting or unsetting."
	case "answer.unknown-question":
		return "Preserve undeclared held answers for explicit history review; submit only declared local questions."
	case "answer.ambiguous-question":
		return "Repair duplicate variable declarations before editing this question."
	case "answer.wrong-kind":
		return "Submit a value kind matching the declared variable type."
	case "answer.not-an-option":
		return "Submit an exact declared choice option."
	case "answer.pattern-mismatch":
		return "Submit a value matching the declared pattern."
	case "directory.unknown-id":
		return "Submit an exact declared directory target, or leave the binding unset."
	case "plan.duplicate-section-id":
		return "Repair repeated held section IDs before editing."
	case "plan.duplicate-variable-name":
		return "Repair repeated variable declarations within their scope before editing."
	case "plan.duplicate-step-id":
		return "Repair repeated step IDs within their section before editing."
	}
	if strings.HasPrefix(code, "secret.") {
		return "Remove credential-like content; supply only nonsecret shared facts."
	}
	switch stage {
	case "input-authoring", "final-authoring":
		return "Repair the indicated held declaration before editing."
	case "input-plan", "final-plan":
		return "Repair the plan's reported structure or stored facts before editing."
	case "submit":
		return "Correct the scoped submission before applying this batch."
	case "switch":
		return "Select a valid held section switch before applying this batch."
	default:
		return "Repair the indicated new instance or ask the catalog maintainer."
	}
}
