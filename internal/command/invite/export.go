package invite

import (
	"slices"

	"onboard-kit/internal/model"
)

const (
	codeWritten        = "written"
	codeCleanupWritten = "cleanup-written"
)

type Result struct {
	Path     string
	Code     string
	Problems []model.Problem
	Blockers []model.Blocker
}

func (result Result) Written() bool {
	return result.Code == codeWritten || result.Code == codeCleanupWritten
}

func (result Result) diagnostics() diagnostics {
	return diagnostics{code: result.Code, problems: result.Problems, blockers: result.Blockers}
}

func refused(report diagnostics) Result {
	return Result{Code: report.code, Problems: report.problems, Blockers: report.blockers}
}

func Export(request Request) Result {
	if !validAccount(request.Account) {
		return refused(diagnostics{code: "account"})
	}
	output, code := openDestination(request.Output)
	if code != "" {
		return refused(diagnostics{code: code})
	}
	defer output.close()
	planSource, preparationSource, code := readSources(request)
	if code != "" {
		return refused(diagnostics{code: code})
	}
	archive, report := buildArchive(request, planSource, preparationSource)
	if report.code != "" {
		return refused(report)
	}
	publication := output.publish(archive)
	if !publication.committed {
		report.code = publication.code
		return refused(report)
	}
	report.code = codeWritten
	if publication.code != "" {
		report.code = publication.code
	}
	result := refused(report)
	result.Path = request.Output
	return result
}

func readSources(request Request) ([]byte, []byte, string) {
	plan, ok := readSource(request.Plan)
	if !ok {
		return nil, nil, "plan-read"
	}
	preparation, ok := readSource(request.Preparation)
	if !ok {
		return nil, nil, "preparation-read"
	}
	return plan, preparation, ""
}

type Readiness struct {
	Code              string
	Plan              model.Plan
	Preparation       model.Preparation
	PreparationErrors []model.Problem
	Blockers          []model.Blocker
	OwnerNamed        bool
	PreaccessContacts []string
}

func (readiness Readiness) Ready() bool {
	return readiness.Code == "" && len(readiness.PreparationErrors) == 0 && len(readiness.Blockers) == 0 && readiness.OwnerNamed
}

func Check(request Request) Readiness {
	if !validAccount(request.Account) {
		return Readiness{Code: "account"}
	}
	planSource, preparationSource, code := readSources(request)
	if code != "" {
		return Readiness{Code: code}
	}
	sources, report := decoded(planSource, preparationSource)
	if report.code != "" {
		return Readiness{Code: report.code}
	}
	readiness := Readiness{
		Plan: sources.plan, Preparation: sources.preparation,
		Blockers: sources.onboarderBlockers(), OwnerNamed: sources.ownerNamed(),
	}
	for _, problem := range sources.preparationProblems(request.Account) {
		if problem.Severity == model.SeverityError {
			readiness.PreparationErrors = append(readiness.PreparationErrors, problem)
		}
	}
	for _, contact := range sources.plan.Directory.Contacts {
		if namedContact(sources.plan.Directory, contact.ID, true) && !slices.Contains(readiness.PreaccessContacts, contact.ID) {
			readiness.PreaccessContacts = append(readiness.PreaccessContacts, contact.ID)
		}
	}
	return readiness
}
