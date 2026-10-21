package validator

import (
	"strings"

	"onboard-kit/internal/model"
)

func contactProblems(directory model.Directory) []model.Problem {
	problems := []model.Problem{}
	for _, contact := range directory.Contacts {
		detail := map[string]string{"contact": contact.ID}
		if isMalformedEmail(contact.Email) {
			problems = append(problems, model.Problem{
				Code:     "directory.email-malformed",
				Severity: model.SeverityWarning,
				Message:  "This email has no @ or has a space.",
				Detail:   detail,
			})
		}
		if isUnreachable(contact) {
			problems = append(problems, model.Problem{
				Code:     "directory.contact-unreachable",
				Severity: model.SeverityWarning,
				Message:  "This contact has no email, slack, or slack id.",
				Detail:   detail,
			})
		}
	}
	return problems
}

func isMalformedEmail(email string) bool {
	return email != "" && (!strings.Contains(email, "@") || strings.Contains(email, " "))
}

func isUnreachable(contact model.Contact) bool {
	return contact.Email == "" && contact.Slack == "" && contact.SlackID == ""
}
