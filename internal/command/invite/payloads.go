package invite

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
	preparationcontract "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/firststeps/assembler"
	invitationencoder "onboard-kit/internal/invitation/encoder"
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	planencoder "onboard-kit/internal/plan/encoder"
	planvalidator "onboard-kit/internal/plan/validator"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
	preparationencoder "onboard-kit/internal/preparation/encoder"
	preparationvalidator "onboard-kit/internal/preparation/validator"
	"onboard-kit/internal/question/finder"
	recordencoder "onboard-kit/internal/record/encoder"
)

func buildArchive(options arguments, planSource, preparationSource []byte) ([]byte, diagnostics) {
	var report diagnostics
	plan, problems := plandecoder.New().Decode(planSource)
	if len(problems) > 0 {
		return nil, report.failed("plan-decode", problems)
	}
	preparation, problems := preparationdecoder.New().Decode(preparationSource)
	if len(problems) > 0 {
		return nil, report.failed("preparation-decode", problems)
	}
	if plan.ID == "" {
		return nil, report.failed("plan-identity", nil)
	}
	problems = planvalidator.New().Validate(plan)
	if hasError(problems) {
		return nil, report.failed("plan-invalid", problems)
	}
	report.problems = append(report.problems, problems...)
	problems = preparationvalidator.New().Validate(preparationContext(options.account, plan), preparation)
	if hasError(problems) {
		return nil, report.failed("preparation-invalid", problems)
	}
	report.problems = append(report.problems, problems...)
	if blockers := finder.New().Blockers(plan, model.AudienceOnboarder); len(blockers) > 0 {
		report.blockers = blockers
		return nil, report.failed("onboarder-blocked", nil)
	}
	if !namedContact(plan.Directory, preparation.Recipient.OnboardingOwner, false) {
		return nil, report.failed("owner", nil)
	}
	if !namedContact(plan.Directory, options.preaccess, true) {
		return nil, report.failed("preaccess-help", nil)
	}
	guide, problems := assembler.New().Assemble(firststepscontract.Context{Account: options.account, GeneralHelp: options.help}, plan, preparation)
	if hasError(problems) {
		return nil, report.failed("firststeps", problems)
	}
	report.problems = append(report.problems, problems...)
	planBytes, problems := planencoder.New().Encode(plan)
	if len(problems) > 0 {
		return nil, report.failed("plan-encode", problems)
	}
	preparationBytes, problems := preparationencoder.New().Encode(preparation)
	if len(problems) > 0 {
		return nil, report.failed("preparation-encode", problems)
	}
	digest := sha256.Sum256(planBytes)
	seed := model.Record{
		SchemaVersion: 1, Account: options.account,
		Plan: model.PlanReference{ID: plan.ID, URL: plan.URL, Hash: hex.EncodeToString(digest[:])},
	}
	recordBytes, problems := recordencoder.New().Encode(seed)
	if len(problems) > 0 {
		return nil, report.failed("record-encode", problems)
	}
	archive, problems := invitationencoder.New().Encode(model.InvitationFiles{
		Account: options.account, Plan: planBytes, Preparation: preparationBytes,
		Record: recordBytes, FirstSteps: []byte(guide),
	})
	if len(problems) > 0 {
		return nil, report.failed("archive", problems)
	}
	return archive, report
}

func (report diagnostics) failed(code string, problems []model.Problem) diagnostics {
	report.code = code
	report.problems = append(report.problems, problems...)
	return report
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func preparationContext(account string, plan model.Plan) preparationcontract.Context {
	context := preparationcontract.Context{Account: account, PlanID: plan.ID}
	for _, contact := range plan.Directory.Contacts {
		context.Contacts = append(context.Contacts, contact.ID)
	}
	for _, team := range plan.Directory.Teams {
		context.Teams = append(context.Teams, team.ID)
	}
	for _, section := range plan.Sections {
		context.Sections = append(context.Sections, section.ID)
	}
	return context
}

func namedContact(directory model.Directory, id string, route bool) bool {
	if id == "" {
		return false
	}
	for _, contact := range directory.Contacts {
		if contact.ID == id {
			return strings.TrimFunc(contact.Name, unicode.IsSpace) != "" && (!route || strings.TrimFunc(contact.Slack, unicode.IsSpace) != "" || strings.TrimFunc(contact.Email, unicode.IsSpace) != "")
		}
	}
	return false
}
