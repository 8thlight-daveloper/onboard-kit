package materializer

import (
	"sort"

	contract "onboard-kit/contracts/catalog/materializer"
	readercontract "onboard-kit/contracts/catalog/reader"
)

func collectGuidance(catalog readercontract.Reader, instances []contract.Instance) ([]contract.Guidance, []contract.Diagnostic) {
	reasons := make(map[string]map[string]bool)
	for _, instance := range instances {
		guides := append([]string{"invitation-prepare"}, additionalGuides(instance.SourceID)...)
		for _, guide := range guides {
			if reasons[guide] == nil {
				reasons[guide] = make(map[string]bool)
			}
			reasons[guide][instance.PermanentSectionID] = true
		}
	}
	ids := make([]string, 0, len(reasons))
	for id := range reasons {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]contract.Guidance, 0, len(ids))
	var diagnostics []contract.Diagnostic
	for _, id := range ids {
		entry, problems := catalog.Read(id)
		if hasErrors(problems) {
			return nil, wrap("guidance", "guidance", problems)
		}
		if entry.Summary.Carrier != readercontract.CarrierPreparationGuidance {
			return nil, []contract.Diagnostic{own("guide-carrier", "guidance", "guidance", "Repair the catalog preparation guide classification.")}
		}
		diagnostics = append(diagnostics, wrap("guidance", "guidance", problems)...)
		instanceIDs := make([]string, 0, len(reasons[id]))
		for instanceID := range reasons[id] {
			instanceIDs = append(instanceIDs, instanceID)
		}
		sort.Strings(instanceIDs)
		markdown := make([]byte, len(entry.Content))
		copy(markdown, entry.Content)
		result = append(result, contract.Guidance{ID: id, InstanceIDs: instanceIDs, Markdown: markdown})
	}
	return result, diagnostics
}

func additionalGuides(source string) []string {
	switch source {
	case "artifact-access", "ci-access", "design-workspace", "documentation-access", "environment-access", "hosting-access", "internal-network", "observability-access", "onboarding-repository", "repository-access", "repository-clone", "secrets-access", "source-host-account", "source-host-membership", "ticket-system-access":
		return []string{"resource-access-arrange"}
	case "company-calendar":
		return []string{"calendar-arrange", "meeting-invite"}
	case "company-email", "company-identity":
		return []string{"identity-arrange"}
	case "day-ninety-review", "day-thirty-review", "onboarding-buddy-meeting", "onboarding-owner-meeting", "week-one-check-in":
		return []string{"meeting-invite", "support-assign"}
	case "deploy-lead":
		return []string{"deploy-support-assign"}
	case "deploy-shadow":
		return []string{"deploy-lead-assign"}
	case "design-contribution", "incident-orientation", "onboarding-help":
		return []string{"support-assign"}
	case "document-reading":
		return []string{"reading-assign"}
	case "office-access":
		return []string{"office-arrange"}
	case "onboarding-welcome":
		return []string{"office-arrange", "remote-arrange", "support-assign"}
	case "people-map", "regular-meetings", "retrospective-participation", "teammate-introduction":
		return []string{"meeting-invite"}
	case "pr-merge":
		return []string{"pr-merge-arrange"}
	case "pr-post":
		return []string{"pr-reviewer-assign"}
	case "pr-review":
		return []string{"pr-review-arrange"}
	case "team-chat":
		return []string{"chat-arrange"}
	case "ticket-ownership":
		return []string{"ticket-assign"}
	case "ticket-pairing":
		return []string{"pair-assign"}
	case "video-meetings":
		return []string{"video-arrange"}
	case "working-device":
		return []string{"device-arrange"}
	case "application-local", "codebase-orientation", "contribution-practices", "design-handoff", "design-system-review", "knowledge-transfer", "onboarding-correction", "onboarding-refresh", "product-orientation", "project-dependencies", "project-tests", "research-review", "security-orientation", "tool-feedback", "working-agreements", "working-hours":
		return nil
	}
	return nil
}
