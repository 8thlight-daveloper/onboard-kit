package materializer

import (
	"sort"
	"strconv"

	contract "onboard-kit/contracts/catalog/materializer"
	readercontract "onboard-kit/contracts/catalog/reader"
	"onboard-kit/internal/catalog/reader"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/authoringvalidator"
	"onboard-kit/internal/plan/validator"
	"onboard-kit/internal/section/hasher"
	"onboard-kit/internal/section/loader"
)

type materializer struct {
	catalog       readercontract.Reader
	sectionLoader loader.Loader
	sectionHasher hasher.Hasher
	author        authoringvalidator.Validator
	planValidator validator.Validator
}

func New() contract.Materializer {
	return materializer{catalog: reader.New(), sectionLoader: loader.New(), sectionHasher: hasher.New(), author: authoringvalidator.New(), planValidator: validator.New()}
}

func (creator materializer) Create(request contract.Request) (contract.Result, []contract.Diagnostic) {
	if diagnostic := admit(request); diagnostic != nil {
		return contract.Result{}, []contract.Diagnostic{*diagnostic}
	}
	if request.PlanID == "" {
		return refuse("plan-id", "request", "planId", "Supply an explicit nonempty plan identity.")
	}
	seen := make(map[string]bool)
	for i, instance := range request.Instances {
		path := instancePath(i) + ".permanentSectionId"
		if !sectionID(instance.PermanentSectionID) {
			return refuse("instance-id", "request", path, "Supply a nonreserved lowercase permanent section ID.")
		}
		if seen[instance.PermanentSectionID] {
			return refuse("duplicate-instance", "request", path, "Give each independent activity its own permanent section ID.")
		}
		seen[instance.PermanentSectionID] = true
	}
	for _, scope := range bindingScopes(request) {
		seenNames := make(map[string]bool)
		for j, binding := range scope.bindings {
			if seenNames[binding.Name] {
				return refuse("duplicate-binding", "request", bindingPath(scope.path, j)+".name", "Supply each scoped binding name once.")
			}
			seenNames[binding.Name] = true
		}
	}
	catalog := creator.catalog
	entries := make([]readercontract.Entry, len(request.Instances))
	for i, instance := range request.Instances {
		entry, problems := catalog.Read(instance.SourceID)
		if len(problems) != 0 {
			return contract.Result{}, wrap("source", instancePath(i)+".sourceId", problems)
		}
		if entry.Summary.Carrier != readercontract.CarrierSectionTemplate {
			return refuse("carrier", "source", instancePath(i)+".sourceId", "Select a section-template source for an activity.")
		}
		entries[i] = entry
	}
	definitions := make([]model.Section, len(entries))
	sectionLoader := creator.sectionLoader
	for i, entry := range entries {
		section, problems := sectionLoader.Load(entry.Content)
		path := instancePath(i) + ".sourceId"
		if len(problems) != 0 {
			return contract.Result{}, wrap("load", path, problems)
		}
		if section.ID != request.Instances[i].SourceID {
			return refuse("source-identity", "profile", path, "Repair the catalog definition to match its declared source identity.")
		}
		if !graphFree(section) {
			return refuse("profile", "profile", path, "Select a graph-free definition until explicit instance mapping is supported.")
		}
		definitions[i] = section
	}
	sectionHasher := creator.sectionHasher
	for i := range definitions {
		definitions[i].ID = request.Instances[i].PermanentSectionID
		baseline := sectionHasher.Hash(definitions[i])
		definitions[i].Source = &model.Source{Template: request.Instances[i].SourceID, Hash: baseline}
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })
	plan := model.Plan{
		SchemaVersion: 1,
		ID:            request.PlanID,
		Provenance:    request.Provenance,
		Directory:     copyDirectory(request.Directory),
		Variables:     copyVariables(request.PlanVariables),
		Sections:      definitions,
		Answers:       model.Answers{Plan: make(map[string]model.Value), Sections: make(map[string]map[string]model.Value)},
	}
	if plan.Variables == nil {
		plan.Variables = make([]model.Variable, 0)
	}
	author := creator.author
	authorProblems := author.Validate(plan)
	if hasErrors(authorProblems) {
		return contract.Result{}, wrap("authoring", "", authorProblems)
	}
	planValidator := creator.planValidator
	if problems := planValidator.Validate(plan); hasErrors(problems) {
		return contract.Result{}, wrap("structure", "", problems)
	}
	scopes := bindingScopes(request)
	for _, scope := range scopes {
		variables := plan.Variables
		if scope.id != "plan" {
			for _, section := range plan.Sections {
				if section.ID == scope.id {
					variables = section.Variables
					break
				}
			}
		}
		for j, binding := range scope.bindings {
			var declaration *model.Variable
			for k := range variables {
				if variables[k].Name == binding.Name {
					declaration = &variables[k]
					break
				}
			}
			if declaration == nil {
				return refuse("binding-name", "binding", bindingPath(scope.path, j)+".name", "Bind an exact variable declared in this scope.")
			}
			if declaration.Audience != "" && declaration.Audience != model.AudienceOnboarder {
				return refuse("binding-owner", "binding", bindingPath(scope.path, j)+".name", "Keep newcomer answers outside the shared organization plan.")
			}
		}
	}
	for _, scope := range scopes {
		answers := plan.Answers.Plan
		if scope.id != "plan" && len(scope.bindings) != 0 {
			answers = make(map[string]model.Value)
			plan.Answers.Sections[scope.id] = answers
		}
		for _, binding := range scope.bindings {
			answers[binding.Name] = copyValue(binding.Value)
		}
	}
	for _, scope := range scopes {
		for j, binding := range scope.bindings {
			problems := planValidator.CheckValue(plan, scope.id+"."+binding.Name, binding.Value)
			if hasErrors(problems) {
				return contract.Result{}, wrap("answer", bindingPath(scope.path, j)+".value", problems)
			}
		}
	}
	planProblems := planValidator.Validate(plan)
	if hasErrors(planProblems) {
		return contract.Result{}, wrap("plan", "", planProblems)
	}
	guidance, diagnostics := collectGuidance(catalog, request.Instances)
	if hasDiagnosticErrors(diagnostics) {
		return contract.Result{}, diagnostics
	}
	success := wrap("authoring", "", authorProblems)
	success = append(success, wrap("plan", "", planProblems)...)
	success = append(success, diagnostics...)
	return contract.Result{Plan: plan, Guidance: guidance}, success
}

type bindingScope struct {
	id       string
	path     string
	bindings []contract.Binding
}

func bindingScopes(request contract.Request) []bindingScope {
	scopes := make([]bindingScope, 0, len(request.Instances)+1)
	scopes = append(scopes, bindingScope{id: "plan", path: "planBindings", bindings: request.PlanBindings})
	for i, instance := range request.Instances {
		scopes = append(scopes, bindingScope{id: instance.PermanentSectionID, path: instancePath(i) + ".bindings", bindings: instance.Bindings})
	}
	return scopes
}

func instancePath(index int) string {
	return "instances[" + strconv.Itoa(index) + "]"
}

func bindingPath(path string, index int) string {
	return path + "[" + strconv.Itoa(index) + "]"
}

func sectionID(id string) bool {
	if id == "" {
		return false
	}
	for i, character := range []byte(id) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || i > 0 && character == '-' {
			continue
		}
		return false
	}
	switch id {
	case "plan", "yes", "no", "y", "n", "true", "false", "on", "off", "null":
		return false
	}
	return true
}

func graphFree(section model.Section) bool {
	if section.Parent != "" || len(section.Requires) != 0 || len(section.After) != 0 || section.IncludeIf != nil || section.Source != nil || section.Enabled != nil || len(section.UnknownFields) != 0 {
		return false
	}
	for _, variable := range section.Variables {
		if variable.AskIf != nil {
			return false
		}
	}
	return true
}

func own(code, stage, path, message string) contract.Diagnostic {
	return contract.Diagnostic{Stage: stage, Path: path, Problem: model.Problem{Code: "materialize." + code, Severity: model.SeverityError, Message: message, Detail: map[string]string{"path": path}}}
}

func refuse(code, stage, path, message string) (contract.Result, []contract.Diagnostic) {
	return contract.Result{}, []contract.Diagnostic{own(code, stage, path, message)}
}

func wrap(stage, path string, problems []model.Problem) []contract.Diagnostic {
	var result []contract.Diagnostic
	for _, problem := range problems {
		if problem.Detail != nil {
			detail := make(map[string]string, len(problem.Detail))
			for key, value := range problem.Detail {
				detail[key] = value
			}
			problem.Detail = detail
		}
		result = append(result, contract.Diagnostic{Stage: stage, Path: path, Problem: problem})
	}
	return result
}

func hasErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func hasDiagnosticErrors(diagnostics []contract.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}
