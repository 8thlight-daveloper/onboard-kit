package assembler

import (
	"strings"

	contract "onboard-kit/contracts/firststeps/assembler"
	locatorcontract "onboard-kit/contracts/preparation/locator"
	"onboard-kit/internal/model"
)

const missingOwner = "Not assigned. Ask the person who sent this invitation to name an onboarding owner and provide a contact route."

func guide(context contract.Context, plan model.Plan, preparation model.Preparation, directory directory, paths locatorcontract.Paths, fragment string) string {
	organization := "Organization not specified."
	if answer := plan.Answers.Plan["organization"]; answer.Kind == model.ValueKindText && answer.Text != "" {
		organization = displayText(answer.Text)
	}
	owner := missingOwner
	if preparation.Recipient.OnboardingOwner != "" {
		owner = contactDisplay(directory.contacts[preparation.Recipient.OnboardingOwner])
	}
	location := "Not specified. Ask the onboarding owner where the shared plan is kept."
	if plan.URL != "" {
		location = displayText(plan.URL)
	}
	generated := "Not specified."
	if plan.Provenance.GeneratedOn != "" {
		generated = displayText(string(plan.Provenance.GeneratedOn))
	}
	if fragment == "" {
		fragment = "No pre-access activities are listed in this plan. Ask the onboarding owner what you need before access works.\n"
	}
	prefix := strings.Join([]string{
		"# First steps",
		"## Welcome",
		"Organization: " + organization,
		"Onboarding owner: " + owner,
		"Use this guide to begin onboarding and keep your own answers, outcomes, and notes in your record.",
		"## What is in the invitation",
		"- `plan.obk.yaml`: shared instructions and organization facts, maintained by onboarders.\n" +
			"- " + displayCode("records/"+context.Account+"/preparation.obk.yaml") + ": arrangements for you, maintained by onboarders.\n" +
			"- " + displayCode("records/"+context.Account+"/record.obk.yaml") + ": your answers, outcomes, and notes, maintained by you.\n" +
			"- `first-steps.md`: this generated guide; it does not replace those three source artifacts.",
		"Invitation account: " + displayCode(context.Account) + ".",
		"Before allowing an import, check that the account is yours. If it is wrong, cancel and ask the sender for a corrected invitation. Do not confirm another person's account.",
		"## Get the tool",
		"Ask the onboarding owner for the supported standalone onboard-kit file and instructions for checking that it runs. A distribution address has not been supplied in this guide. You should not need organization access or another runtime to begin.",
		"## Start",
		"Use the invitation-opening instructions supplied with the supported tool. If none were supplied, ask the onboarding owner before proceeding. This guide does not import files or confirm identity.",
		"An import must show the account and destinations before confirmation and write nothing if you decline. Reading this guide is not recorded first use. The supported flow must explain any first-use record write when you confirm and begin onboarding; first use is not an employment start date.",
		"Relative record destination: " + displayCode(paths.Record) + ".",
		"Relative preparation destination: " + displayCode(paths.Preparation) + ".",
		"These destinations are relative to the root explicitly selected by the import flow. Archive names are not extraction instructions. Existing personal work must be preserved; ask the owner for help with a conflict.",
		"## Do these first",
	}, "\n\n")
	suffix := strings.Join([]string{
		"## Connect to the shared repository",
		"Plan location: " + location,
		"After access works, ask the onboarding owner for the repository and supported sharing instructions. Check the record destination above before adding, committing, or pushing your own record. Preparation and the shared plan have different owners. No Git operation is performed by this guide.",
		"## Get the latest plan",
		"Once you can reach the shared plan, ask the onboarding owner or consult the supported tool's help for the update workflow. Keep your saved answers, outcomes, notes, and original first-use event. A newer plan format may require a newer supported tool; do not edit the version to bypass that check.",
		"## Day to day",
		"Keep the invitation and your record, and use the supported tool's help to return to onboarding. Verified means the activity's required checks and prerequisites have been met. Failed records an unsuccessful attempt; deferred records a deliberate postponement. An arranged assignment is not a verified activity. Timing windows are planning goals, not failures or proof of access.",
		"## Who can see your record",
		"Anyone with read access to the shared repository can read answers and notes stored there. Separate file ownership does not make personal information private. Do not put secrets or unnecessary personal information in shared files.",
		"## If something goes wrong",
		"- Lost plan: ask the onboarding owner for a fresh invitation or the supported repository recovery workflow.\n" +
			"- Lost record: check the configured record destination and shared repository history with the onboarding owner before creating a replacement.\n" +
			"- Wrong account: stop and ask the sender for a correction. An account change after use needs an explicit supported move that preserves your work.\n" +
			"- Tool will not run: ask the onboarding owner for a supported file and setup help.\n" +
			"- Plan is newer than the tool understands: obtain the supported tool from the onboarding owner; do not change the plan version by hand.\n" +
			"- Access does not work: ask the onboarding owner or the help contacts below. A requested or arranged account is not proof that access works.",
		"## Who to ask",
		"Onboarding owner: " + owner,
		directory.helpList(context.GeneralHelp),
		"## About this plan",
		"Plan ID: " + displayText(plan.ID),
		"Plan generated date: " + generated,
	}, "\n\n")
	return prefix + "\n\n" + fragment + "\n" + suffix + "\n"
}
