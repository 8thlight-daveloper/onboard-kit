package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

const planKind = "plan"

func (c *collector) checkKind(node *yaml.Node, at location) {
	found := c.stringOf(node, at)
	if found == "" || found == planKind {
		return
	}
	c.reportProblem(wrongKindProblem(found, at), at)
}

func wrongKindProblem(found string, at location) model.Problem {
	problem := pathProblem(codeWrongKind, at.path, map[string]string{"found": found, "expected": planKind})
	problem.Message = `the kind "` + found + `" is not a plan, a plan is expected`
	return problem
}
