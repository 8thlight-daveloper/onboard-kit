package decoder

import (
	"regexp"
	"strconv"

	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

const (
	versionKey       = "schemaVersion"
	supportedVersion = 1
)

var canonicalInteger = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

func checkVersion(root *yaml.Node) []model.Problem {
	values := versionValues(root)
	switch len(values) {
	case 0:
		return []model.Problem{pathProblem(codeVersionInvalid, versionKey, nil)}
	case 1:
		return versionValueProblems(values[0])
	}
	return []model.Problem{pathProblem(codeDuplicateKey, versionKey, nil)}
}

func versionValues(root *yaml.Node) []*yaml.Node {
	var values []*yaml.Node
	for index := 0; index+1 < len(root.Content); index += 2 {
		if root.Content[index].Value == versionKey {
			values = append(values, root.Content[index+1])
		}
	}
	return values
}

func versionValueProblems(value *yaml.Node) []model.Problem {
	number, ok := versionNumber(value)
	if !ok {
		return []model.Problem{pathProblem(codeVersionInvalid, versionKey, nil)}
	}
	switch {
	case number < supportedVersion:
		return []model.Problem{versionProblem(codeVersionTooOld, value.Value)}
	case number > supportedVersion:
		return []model.Problem{versionProblem(codeVersionTooNew, value.Value)}
	}
	return nil
}

func versionNumber(value *yaml.Node) (int64, bool) {
	if !isPlain(value) || !canonicalInteger.MatchString(value.Value) {
		return 0, false
	}
	number, err := strconv.ParseInt(value.Value, 10, 64)
	return number, err == nil
}

func versionProblem(code string, found string) model.Problem {
	return pathProblem(code, versionKey, map[string]string{
		"found":     found,
		"supported": strconv.Itoa(supportedVersion),
	})
}
