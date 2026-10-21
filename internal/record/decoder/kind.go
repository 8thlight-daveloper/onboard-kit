package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

const (
	kindKey    = "kind"
	recordKind = "record"
	kindNeeded = "a record file needs kind: record"
)

func checkKind(root *yaml.Node) []model.Problem {
	values := valuesOfKey(root, kindKey)
	switch len(values) {
	case 0:
		return []model.Problem{kindMissingProblem()}
	case 1:
		return kindValueProblems(values[0])
	}
	return []model.Problem{pathProblem(codeDuplicateKey, kindKey, nil)}
}

func kindValueProblems(value *yaml.Node) []model.Problem {
	switch {
	case isNull(value):
		return []model.Problem{kindMissingProblem()}
	case value.Kind != yaml.ScalarNode:
		return []model.Problem{kindNotARecordProblem("", "the file has a kind that is not text. "+kindNeeded)}
	case value.Value != recordKind:
		return []model.Problem{kindNotARecordProblem(value.Value, "the file has kind \""+value.Value+"\". "+kindNeeded)}
	}
	return nil
}

func kindMissingProblem() model.Problem {
	problem := pathProblem(codeKindMissing, kindKey, nil)
	problem.Message = "the file has no kind. " + kindNeeded
	return problem
}

func kindNotARecordProblem(found string, message string) model.Problem {
	problem := pathProblem(codeKindNotARecord, kindKey, map[string]string{"found": found})
	problem.Message = message
	return problem
}
