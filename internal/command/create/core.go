package create

import (
	materializercontract "onboard-kit/contracts/catalog/materializer"
	readercontract "onboard-kit/contracts/catalog/reader"
	"onboard-kit/internal/catalog/materializer"
	"onboard-kit/internal/catalog/reader"
	"onboard-kit/internal/model"
)

type stagedSection struct {
	source   string
	instance string
}

var coreSet = []stagedSection{
	{"onboarding-welcome", "onboarding-welcome"},
	{"onboarding-help", "onboarding-help"},
	{"onboarding-owner-meeting", "onboarding-owner-meeting"},
	{"onboarding-buddy-meeting", "onboarding-buddy-meeting"},
	{"people-map", "people-map"},
	{"working-agreements", "working-agreements"},
	{"regular-meetings", "regular-meetings"},
	{"team-chat", "team-chat"},
	{"standup-attendance", "standup-attendance"},
	{"source-host-account", "source-host-account"},
	{"repository-access", "repository-access"},
	{"ticket-system-access", "ticket-system-access"},
	{"ci-access", "ci-access"},
	{"environment-access", "environment-access"},
	{"documentation-access", "documentation-access"},
	{"project-dependencies", "project-dependencies"},
	{"application-local", "application-local"},
	{"project-tests", "project-tests"},
	{"codebase-tutor", "codebase-tutor"},
	{"contribution-practices", "contribution-practices"},
	{"ticket-pairing", "ticket-pairing-navigator"},
	{"ticket-pairing", "ticket-pairing-driver"},
	{"pr-post", "pr-post"},
	{"ai-tool-policy", "ai-tool-policy"},
	{"deploy-shadow", "deploy-shadow"},
	{"deploy-lead", "deploy-lead"},
	{"retrospective-participation", "retrospective-participation"},
	{"week-one-check-in", "week-one-check-in"},
	{"day-thirty-review", "day-thirty-review"},
	{"day-ninety-review", "day-ninety-review"},
	{"onboarding-correction", "onboarding-correction"},
}

func CorePlan(identifier string, templateSet string, generatedOn string) (model.Plan, []model.Problem) {
	request := materializercontract.Request{
		PlanID:     identifier,
		Provenance: model.Provenance{TemplateSet: templateSet, GeneratedOn: model.Date(generatedOn)},
		Instances:  coreInstances(),
	}
	result, diagnostics := materializer.New().Create(request)
	var problems []model.Problem
	failed := false
	for _, diagnostic := range diagnostics {
		problems = append(problems, diagnostic.Problem)
		failed = failed || diagnostic.Problem.Severity == model.SeverityError
	}
	if failed {
		return model.Plan{}, problems
	}
	plan := result.Plan
	plan.Variables = append(identityVariables(), plan.Variables...)
	plan.Sections = append([]model.Section{planSourceSection()}, plan.Sections...)
	return plan, nil
}

func identityVariables() []model.Variable {
	return []model.Variable{
		{
			Name:        "organization",
			Description: "What is the organization's name?",
			Example:     "ACME Corp",
			Required:    true,
			Audience:    model.AudienceOnboarder,
		},
	}
}

func planSourceSection() model.Section {
	return model.Section{
		ID:      "plan-source",
		Title:   "Where this plan is kept",
		Summary: "Name the repository and the file that hold this plan",
		Timing:  model.TimingBeforeDayOne,
		Body: "This plan is kept in $plan_repository, in the file $plan_file.\n\n" +
			"Get updates to it from that repository.\n",
		Variables: []model.Variable{
			{
				Name:        "plan_repository",
				Description: "What is the address of the repository that holds the plan?",
				Example:     "https://example.test/acme/onboarding.git",
				Audience:    model.AudienceOnboarder,
			},
			{
				Name:        "plan_file",
				Description: "What is the relative path of the plan file in the repository?",
				Example:     "plans/platform.obk.yaml",
				Audience:    model.AudienceOnboarder,
			},
		},
	}
}

func coreInstances() []materializercontract.Instance {
	available := map[string]bool{}
	for _, summary := range reader.New().List() {
		available[summary.ID] = summary.Carrier == readercontract.CarrierSectionTemplate
	}
	instances := make([]materializercontract.Instance, 0, len(coreSet))
	for _, staged := range coreSet {
		if available[staged.source] {
			instances = append(instances, materializercontract.Instance{SourceID: staged.source, PermanentSectionID: staged.instance})
		}
	}
	return instances
}
