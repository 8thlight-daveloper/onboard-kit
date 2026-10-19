package hasher

import (
	"sort"

	"onboard-kit/internal/model"
)

func stepsText(steps []model.Step) string {
	items := make([]string, len(steps))
	for index, step := range steps {
		items[index] = stepText(step)
	}
	return listText(items)
}

func stepText(step model.Step) string {
	return fields{}.
		withString("id", step.ID).
		withString("name", step.Name).
		withProse("notes", step.Notes).
		withString("owner", step.Owner).
		withString("timing", string(step.Timing)).
		withStrings("escalation", step.Escalation).
		withStrings("run", step.Run).
		withString("open", step.Open).
		withString("directory", step.Directory).
		withRaw("env", environmentText(step.Env), len(step.Env) > 0).
		withRaw("checks", checksText(step.Checks), len(step.Checks) > 0).
		text()
}

func environmentText(environment map[string]string) string {
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	members := make(fields, 0, len(keys))
	for _, key := range keys {
		members = members.withRaw(key, quote(environment[key]), true)
	}
	return members.text()
}

func checksText(checks []model.Check) string {
	items := make([]string, len(checks))
	for index, check := range checks {
		items[index] = checkText(check)
	}
	return listText(items)
}

func checkText(check model.Check) string {
	return fields{}.
		withStrings("run", check.Run).
		withProse("ask", check.Ask).
		withString("env", check.Env).
		text()
}
