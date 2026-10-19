package loader

import (
	"sort"

	"onboard-kit/internal/model"
)

const (
	codeNotYAML           = "file.not-yaml"
	codeNotATemplate      = "file.not-a-template"
	codeMultipleDocuments = "file.multiple-documents"
	codeVersionInvalid    = "version.invalid"
	codeVersionTooOld     = "version.too-old"
	codeVersionTooNew     = "version.too-new"
	codeBooleanSpelling   = "format.boolean-spelling"
	codeNumberSpelling    = "format.number-spelling"
	codeNullSpelling      = "format.null-spelling"
	codeAlias             = "format.alias"
	codeDuplicateKey      = "format.duplicate-key"
	codeWrongType         = "format.wrong-type"
	codeUnknownField      = "format.unknown-field"
	codeEmptyAnswer       = "format.empty-answer"
	codeEmptyItem         = "format.empty-item"
)

var codeOrder = []string{
	codeBooleanSpelling,
	codeNumberSpelling,
	codeNullSpelling,
	codeAlias,
	codeDuplicateKey,
	codeWrongType,
	codeUnknownField,
	codeEmptyAnswer,
	codeEmptyItem,
}

var messages = map[string]string{
	codeNotYAML:           "the file is not valid YAML in UTF-8",
	codeNotATemplate:      "the file is empty or its top level is not a map",
	codeMultipleDocuments: "the file holds more than one YAML document",
	codeVersionInvalid:    "schemaVersion is missing or is not a plain integer",
	codeVersionTooOld:     "schemaVersion is older than this tool supports",
	codeVersionTooNew:     "schemaVersion is newer than this tool supports, upgrade the tool",
	codeBooleanSpelling:   "write the boolean as true or false",
	codeNumberSpelling:    "write the number in plain canonical form, or quote the text",
	codeNullSpelling:      "write null as null or leave the value empty",
	codeAlias:             "anchors, aliases, and merge keys are not allowed",
	codeDuplicateKey:      "the key is written twice in one map",
	codeWrongType:         "the value has the wrong kind for this field",
	codeUnknownField:      "the field is not known here",
	codeEmptyAnswer:       "an answer cannot be null",
	codeEmptyItem:         "a list item cannot be null",
}

type finding struct {
	problem model.Problem
	line    int
	column  int
	rank    int
}

type collector struct {
	findings []finding
}

func (c *collector) report(code string, at location) {
	c.findings = append(c.findings, finding{
		problem: pathProblem(code, at.path, nil),
		line:    at.line,
		column:  at.column,
		rank:    rankOf(code),
	})
}

func (c *collector) sortedProblems() []model.Problem {
	ordered := append([]finding(nil), c.findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return before(ordered[i], ordered[j])
	})
	problems := make([]model.Problem, len(ordered))
	for index, entry := range ordered {
		problems[index] = entry.problem
	}
	return problems
}

func before(left finding, right finding) bool {
	if left.line != right.line {
		return left.line < right.line
	}
	if left.column != right.column {
		return left.column < right.column
	}
	return left.rank < right.rank
}

func rankOf(code string) int {
	for index, candidate := range codeOrder {
		if candidate == code {
			return index
		}
	}
	return len(codeOrder)
}

func fileProblem(code string) model.Problem {
	return model.Problem{Code: code, Severity: model.SeverityError, Message: messages[code]}
}

func pathProblem(code string, path string, extra map[string]string) model.Problem {
	detail := map[string]string{"path": path}
	for key, value := range extra {
		detail[key] = value
	}
	return model.Problem{Code: code, Severity: model.SeverityError, Message: messages[code], Detail: detail}
}
