package encoder

import "onboard-kit/internal/model"

func (c *collector) stepSequence(steps []model.Step, path string) node {
	return sequenceOf(steps, func(index int, step model.Step) node {
		return c.stepNode(step, indexed(path, index))
	})
}

func (c *collector) stepNode(step model.Step, path string) node {
	builder := &mappingBuilder{}
	builder.add("id", stringField(step.ID))
	builder.add("name", stringField(step.Name))
	builder.add("notes", stringField(step.Notes))
	builder.add("owner", stringField(step.Owner))
	builder.add("timing", stringField(string(step.Timing)))
	builder.add("escalation", listField(step.Escalation))
	builder.add("run", listField(step.Run))
	builder.add("open", stringField(step.Open))
	builder.add("directory", stringField(step.Directory))
	builder.add("env", environmentNode(step.Env))
	builder.add("checks", c.checkSequence(step.Checks, path+".checks"))
	return c.requireKeys(builder.build(), path)
}

func environmentNode(environment map[string]string) node {
	builder := &mappingBuilder{}
	for _, name := range sortedKeys(environment) {
		builder.add(name, textNode(environment[name]))
	}
	return builder.build()
}

func (c *collector) checkSequence(checks []model.Check, path string) node {
	return sequenceOf(checks, func(index int, check model.Check) node {
		return c.checkNode(check, indexed(path, index))
	})
}

func (c *collector) checkNode(check model.Check, path string) node {
	if countKinds(check) != 1 {
		c.report(codeAmbiguousCheck, path, "a check must have exactly one of run, ask, or env")
		return nil
	}
	builder := &mappingBuilder{}
	builder.add("run", listField(check.Run))
	builder.add("ask", stringField(check.Ask))
	builder.add("env", stringField(check.Env))
	return builder.build()
}

func countKinds(check model.Check) int {
	count := 0
	if len(check.Run) > 0 {
		count++
	}
	if check.Ask != "" {
		count++
	}
	if check.Env != "" {
		count++
	}
	return count
}
