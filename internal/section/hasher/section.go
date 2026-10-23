package hasher

import "onboard-kit/internal/model"

func sectionText(section model.Section) string {
	timing := section.Timing
	if timing == "" {
		timing = model.DefaultTiming
	}
	return fields{}.
		withString("id", section.ID).
		withProse("title", section.Title).
		withProse("summary", section.Summary).
		withRaw("includeIf", conditionTextOrEmpty(section.IncludeIf), hasClauses(section.IncludeIf)).
		withString("parent", section.Parent).
		withFlag("bootstrap", section.Bootstrap).
		withString("timing", string(timing)).
		withRaw("after", quotedSetText(section.After), len(section.After) > 0).
		withRaw("requires", identifierSetText(section.Requires), len(section.Requires) > 0).
		withString("owner", section.Owner).
		withStrings("escalation", section.Escalation).
		withProse("verification", section.Verification).
		withProse("body", section.Body).
		withRaw("variables", variablesText(section.Variables), len(section.Variables) > 0).
		withRaw("steps", stepsText(section.Steps), len(section.Steps) > 0).
		withRaw("unknownFields", mapText(section.UnknownFields), len(section.UnknownFields) > 0).
		text()
}
