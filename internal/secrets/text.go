package secrets

import "onboard-kit/internal/model"

type shape struct {
	code    string
	message string
	matches func(text string) bool
}

var shapes = [...]shape{
	{
		code:    "secret.private-key",
		message: "This value looks like a private key. Keep private keys out of the plan.",
		matches: hasPrivateKeyHeader,
	},
	{
		code:    "secret.url-credential",
		message: "This value looks like a URL with a password in it. Keep passwords out of the plan.",
		matches: hasURLPassword,
	},
	{
		code:    "secret.token-prefix",
		message: "This value looks like an access token. Keep tokens out of the plan.",
		matches: hasTokenPrefix,
	},
	{
		code:    "secret.random-string",
		message: "This value contains a long random-looking string that may be a secret. Keep secrets out of the plan.",
		matches: hasRandomString,
	},
}

func textProblems(text string, detail map[string]string) []model.Problem {
	for _, candidate := range shapes {
		if candidate.matches(text) {
			return []model.Problem{{
				Code:     candidate.code,
				Severity: model.SeverityError,
				Message:  candidate.message,
				Detail:   detail,
			}}
		}
	}
	return nil
}
