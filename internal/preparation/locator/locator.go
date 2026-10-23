package locator

import (
	"strings"
	"unicode"
	"unicode/utf8"

	contract "onboard-kit/contracts/preparation/locator"
	"onboard-kit/internal/model"
)

const (
	accountToken        = "{account}"
	preparationFilename = "preparation.obk.yaml"
	pathBudget          = 4096
)

type locator struct{}

func New() contract.Locator {
	return locator{}
}

func (locator) Locate(resultsPath, account string) (contract.Paths, []model.Problem) {
	if !validAccount(account) {
		return failure("preparation.account", "account", "Use an account beginning with an ASCII letter or digit and containing only ASCII letters, digits, dots, underscores, or hyphens.")
	}
	if resultsPath == "" {
		resultsPath = "records/{account}/record.obk.yaml"
	}
	if !validTemplate(resultsPath) {
		return failure("preparation.results-path", "resultsPath", "Use a relative slash path with printable names, no traversal or empty components, and only the account placeholder.")
	}
	lastSlash := strings.LastIndexByte(resultsPath, '/')
	if lastSlash < 0 || !strings.Contains(resultsPath[:lastSlash], accountToken) {
		return failure("preparation.account-directory", "resultsPath", "Put the literal account placeholder in a directory component so each account has its own preparation sibling.")
	}
	directory := resultsPath[:lastSlash+1]
	if !fitsBudget(resultsPath, len(account), 0) || !fitsBudget(directory, len(account), len(preparationFilename)) {
		return failure("preparation.path-too-long", "resultsPath", "Shorten the results path or account so both expanded paths fit within 4096 UTF-8 bytes.")
	}
	record := strings.ReplaceAll(resultsPath, accountToken, account)
	filename := record[strings.LastIndexByte(record, '/')+1:]
	if asciiEqualFold(filename, preparationFilename) {
		return failure("preparation.path-collision", "resultsPath", "Choose a record filename distinct from the fixed preparation sibling, including ASCII case variants.")
	}
	preparation := record[:strings.LastIndexByte(record, '/')+1] + preparationFilename
	return contract.Paths{Record: record, Preparation: preparation}, nil
}

func validAccount(account string) bool {
	if account == "" || !asciiAlphanumeric(account[0]) {
		return false
	}
	for index := 1; index < len(account); index++ {
		character := account[index]
		if !asciiAlphanumeric(character) && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphanumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9'
}

func validTemplate(template string) bool {
	if !utf8.ValidString(template) {
		return false
	}
	for _, component := range strings.Split(template, "/") {
		if component == "" || component == "." || component == ".." || component[0] == ' ' || component[len(component)-1] == ' ' {
			return false
		}
		literal := strings.ReplaceAll(component, accountToken, "")
		for _, character := range literal {
			if character == '{' || character == '}' || character == ':' || character == '\\' || !unicode.IsPrint(character) || character != ' ' && unicode.IsSpace(character) {
				return false
			}
		}
	}
	return true
}

func fitsBudget(template string, accountLength, suffixLength int) bool {
	count := strings.Count(template, accountToken)
	literalLength := len(template) - count*len(accountToken)
	if literalLength > pathBudget-suffixLength {
		return false
	}
	return count == 0 || accountLength <= (pathBudget-suffixLength-literalLength)/count
}

func asciiEqualFold(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		character := left[index]
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		if character != right[index] {
			return false
		}
	}
	return true
}

func failure(code, path, message string) (contract.Paths, []model.Problem) {
	return contract.Paths{}, []model.Problem{{
		Code:     code,
		Severity: model.SeverityError,
		Message:  message,
		Detail:   map[string]string{"path": path},
	}}
}
