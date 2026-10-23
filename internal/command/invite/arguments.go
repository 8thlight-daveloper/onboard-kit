package invite

import (
	"strings"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
)

const usage = "Usage: onboard-kit invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID [--help-contact ID | --help-role ID ...]\n"

type arguments struct {
	plan        string
	preparation string
	account     string
	output      string
	preaccess   string
	help        []firststepscontract.HelpReference
}

func parse(args []string) (arguments, bool, bool) {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "export" && len(args) > 1 && args[1] == "--help") {
		return arguments{}, true, true
	}
	if len(args) == 0 || args[0] != "export" {
		return arguments{}, false, false
	}
	var result arguments
	scalars := map[string]*string{
		"--plan": &result.plan, "--preparation": &result.preparation,
		"--account": &result.account, "--out": &result.output,
		"--preaccess-help-contact": &result.preaccess,
	}
	seen := make(map[string]bool)
	for index := 1; index < len(args); index++ {
		name, value, equals := strings.Cut(args[index], "=")
		field, scalar := scalars[name]
		if !scalar && name != "--help-contact" && name != "--help-role" {
			return arguments{}, false, false
		}
		if !equals {
			index++
			if index == len(args) || strings.HasPrefix(args[index], "--") {
				return arguments{}, false, false
			}
			value = args[index]
		}
		if value == "" || scalar && seen[name] {
			return arguments{}, false, false
		}
		if scalar {
			*field = value
			seen[name] = true
		}
		switch name {
		case "--preaccess-help-contact", "--help-contact":
			result.help = append(result.help, firststepscontract.HelpReference{Kind: "contact", ID: value})
		case "--help-role":
			result.help = append(result.help, firststepscontract.HelpReference{Kind: "role", ID: value})
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
