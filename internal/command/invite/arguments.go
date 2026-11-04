package invite

import (
	"strings"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
)

const usage = "Usage: onboard-kit invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID [--help-contact ID | --help-role ID ...]\n"

type Request struct {
	Plan        string
	Preparation string
	Account     string
	Output      string
	Preaccess   string
	Help        []firststepscontract.HelpReference
}

func parse(args []string) (Request, bool, bool) {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "export" && len(args) > 1 && args[1] == "--help") {
		return Request{}, true, true
	}
	if len(args) == 0 || args[0] != "export" {
		return Request{}, false, false
	}
	var result Request
	scalars := map[string]*string{
		"--plan": &result.Plan, "--preparation": &result.Preparation,
		"--account": &result.Account, "--out": &result.Output,
		"--preaccess-help-contact": &result.Preaccess,
	}
	seen := make(map[string]bool)
	for index := 1; index < len(args); index++ {
		name, value, equals := strings.Cut(args[index], "=")
		field, scalar := scalars[name]
		if !scalar && name != "--help-contact" && name != "--help-role" {
			return Request{}, false, false
		}
		if !equals {
			index++
			if index == len(args) || strings.HasPrefix(args[index], "--") {
				return Request{}, false, false
			}
			value = args[index]
		}
		if value == "" || scalar && seen[name] {
			return Request{}, false, false
		}
		if scalar {
			*field = value
			seen[name] = true
		}
		switch name {
		case "--preaccess-help-contact", "--help-contact":
			result.Help = append(result.Help, firststepscontract.HelpReference{Kind: "contact", ID: value})
		case "--help-role":
			result.Help = append(result.Help, firststepscontract.HelpReference{Kind: "role", ID: value})
		}
	}
	return result, false, len(seen) == len(scalars)
}

func validAccount(account string) bool {
	if len(account) == 0 || len(account) > 4096-len("records/")-len("/preparation.obk.yaml") || !alphanumeric(account[0]) {
		return false
	}
	for index := 1; index < len(account); index++ {
		character := account[index]
		if !alphanumeric(character) && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func alphanumeric(character byte) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9'
}
