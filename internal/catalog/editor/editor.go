package editor

import (
	contract "onboard-kit/contracts/catalog/editor"
	creation "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/catalog/materializer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/authoringvalidator"
	planeditor "onboard-kit/internal/plan/editor"
	"onboard-kit/internal/plan/validator"
	"reflect"
	"strconv"
)

type editor struct {
	creation  creation.Materializer
	changing  planeditor.Editor
	authoring authoringvalidator.Validator
	checking  validator.Validator
}

func New() contract.Editor {
	return editor{creation: materializer.New(), changing: planeditor.New(), authoring: authoringvalidator.New(), checking: validator.New()}
}

func (editing editor) Edit(plan model.Plan, operations []contract.Operation) (model.Plan, []contract.Diagnostic) {
	candidate, copied, code := admit(plan, operations)
	if code != "" {
		return refuse(own(code, "admission", admissionPath(code)))
	}
	if plan.ID == "" {
		return refuse(own("edit.plan-id", "admission", "plan.id"))
	}
	if problems := editing.authoring.Validate(plan); hasErrors(problems) {
		return model.Plan{}, project(problems, "input-authoring", "plan")
	}
	if problems := editing.checking.Validate(plan); hasErrors(problems) {
		return model.Plan{}, project(problems, "input-plan", "plan")
	}
	reserved := make(map[string]bool)
	for _, section := range plan.Sections {
		reserved[section.ID] = true
	}
	for id := range plan.Answers.Sections {
		reserved[id] = true
	}
	for index, operation := range copied {
		path := operationPath(index)
		if !validOperation(operation) {
			return refuse(own("edit.operation", "operation", path))
		}
		if operation.Kind == contract.Submit && !validKind(operation.Value.Kind) {
			return refuse(own("edit.value-kind", "operation", path+".value"))
		}
		if operation.Kind == contract.Add {
			if reserved[operation.SectionID] {
				return refuse(own("edit.reserved-id", "reservation", path+".sectionId"))
			}
			reserved[operation.SectionID] = true
		}
	}
	var warnings []contract.Diagnostic
	for index, operation := range copied {
		updated, diagnostics := editing.apply(candidate, operation, operationPath(index))
		if diagnosticErrors(diagnostics) {
			return model.Plan{}, diagnostics
		}
		candidate = updated
		warnings = append(warnings, diagnostics...)
	}
	result, _, code := admit(candidate, nil)
	if code != "" {
		return refuse(own(code, "admission", admissionPath(code)))
	}
	problems := editing.authoring.Validate(result)
	diagnostics := project(problems, "final-authoring", "plan")
	if hasErrors(problems) {
		return model.Plan{}, diagnostics
	}
	warnings = append(warnings, diagnostics...)
	problems = editing.checking.Validate(result)
	diagnostics = project(problems, "final-plan", "plan")
	if hasErrors(problems) {
		return model.Plan{}, diagnostics
	}
	warnings = append(warnings, diagnostics...)
	return result, warnings
}

func validOperation(operation contract.Operation) bool {
	zeroValue := reflect.DeepEqual(operation.Value, model.Value{})
	switch operation.Kind {
	case contract.Add:
		return operation.Name == "" && zeroValue
	case contract.Submit:
		return operation.SourceID == "" && len(operation.Bindings) == 0
	case contract.Unset:
		return operation.SourceID == "" && len(operation.Bindings) == 0 && zeroValue
	case contract.Retire, contract.Reactivate:
		return operation.SourceID == "" && len(operation.Bindings) == 0 && operation.Name == "" && zeroValue
	default:
		return false
	}
}

func validKind(kind model.ValueKind) bool {
	return kind == model.ValueKindNone || kind == model.ValueKindText || kind == model.ValueKindFlag || kind == model.ValueKindList || kind == model.ValueKindDeferred
}

func operationPath(index int) string {
	return "operations[" + strconv.Itoa(index) + "]"
}

func (editing editor) apply(plan model.Plan, operation contract.Operation, path string) (model.Plan, []contract.Diagnostic) {
	if operation.Kind == contract.Add {
		return editing.add(plan, operation, path)
	}
	sectionIndex := -1
	for index, section := range plan.Sections {
		if section.ID == operation.SectionID {
			sectionIndex = index
			break
		}
	}
	if sectionIndex < 0 {
		return refuse(own("edit.section", "operation", path+".sectionId"))
	}
	if operation.Kind == contract.Retire || operation.Kind == contract.Reactivate {
		updated, problems := editing.changing.SetSectionEnabled(plan, operation.SectionID, operation.Kind == contract.Reactivate)
		return updated, project(problems, "switch", path)
	}
	var declaration *model.Variable
	for index := range plan.Sections[sectionIndex].Variables {
		variable := &plan.Sections[sectionIndex].Variables[index]
		if variable.Name == operation.Name {
			declaration = variable
			break
		}
	}
	if declaration == nil {
		return refuse(own("edit.binding", "operation", path+".name"))
	}
	if declaration.Audience != "" && declaration.Audience != model.AudienceOnboarder {
		return refuse(own("edit.owner", "operation", path+".name"))
	}
	value := operation.Value
	if operation.Kind == contract.Unset {
		value = model.Value{}
	}
	before := len(plan.Answers.Sections[operation.SectionID])
	updated, problems := editing.changing.Submit(plan, operation.SectionID+"."+operation.Name, value)
	if !hasErrors(problems) && before > 0 && len(updated.Answers.Sections[operation.SectionID]) == 0 {
		delete(updated.Answers.Sections, operation.SectionID)
	}
	return updated, project(problems, "submit", path)
}

func (editing editor) add(plan model.Plan, operation contract.Operation, path string) (model.Plan, []contract.Diagnostic) {
	bindings := make([]creation.Binding, len(operation.Bindings))
	if operation.Bindings == nil {
		bindings = nil
	}
	for index, binding := range operation.Bindings {
		bindings[index] = creation.Binding{Name: binding.Name, Value: binding.Value}
	}
	directory := clone(reflect.ValueOf(plan.Directory)).Interface().(model.Directory)
	result, problems := editing.creation.Create(creation.Request{
		PlanID:    plan.ID,
		Directory: directory,
		Instances: []creation.Instance{{SourceID: operation.SourceID, PermanentSectionID: operation.SectionID, Bindings: bindings}},
	})
	var diagnostics []contract.Diagnostic
	for _, problem := range problems {
		location := path
		if problem.Path != "" {
			location += "." + problem.Path
		}
		diagnostics = append(diagnostics, diagnostic(problem.Problem, "add-"+problem.Stage, location))
	}
	if diagnosticErrors(diagnostics) {
		return model.Plan{}, diagnostics
	}
	plan.Sections = append(plan.Sections, result.Plan.Sections...)
	if scope := result.Plan.Answers.Sections[operation.SectionID]; len(scope) != 0 {
		if plan.Answers.Sections == nil {
			plan.Answers.Sections = make(map[string]map[string]model.Value)
		}
		plan.Answers.Sections[operation.SectionID] = scope
	}
	return plan, diagnostics
}
