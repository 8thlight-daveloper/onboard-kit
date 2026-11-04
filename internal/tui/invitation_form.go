package tui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	reviewercontract "onboard-kit/contracts/preparation/reviewer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/preparation/reviewer"
)

const (
	invitationPreaccessField = iota
	invitationHelpField
	invitationOutputField
	invitationFieldCount
)

const (
	helpKindContact = "contact"
	helpKindRole    = "role"
	helpValueMark   = ":"
)

var invitationFieldLabels = [invitationFieldCount]string{"Pre-access help", "General help", "File"}

type invitationForm struct {
	person        InvitationPerson
	focus         int
	preaccess     string
	help          []string
	output        textinput.Model
	choosing      bool
	chooser       picker
	problems      []string
	guide         string
	guideProblems []model.Problem
	guideScroll   int
}

func newInvitationForm(person InvitationPerson, width int) invitationForm {
	output := newTextField(person.Output)
	output.Placeholder = "e.g. obk-invitation-" + person.Account + ".zip"
	form := invitationForm{person: person, output: output}
	return form.withGuide().withWidth(width).focused(0)
}

func (form invitationForm) withWidth(width int) invitationForm {
	form.output.Width = max(width-formMarkWidth-preparationLabelWidth-1, inputMinimum)
	form.output.SetCursor(form.output.Position())
	return form
}

func (form invitationForm) focused(index int) invitationForm {
	form.focus = max(min(index, invitationFieldCount-1), 0)
	if form.onOutput() {
		form.output.Focus()
	} else {
		form.output.Blur()
	}
	return form
}

func (form invitationForm) onOutput() bool { return form.focus == invitationOutputField }

func helpValue(kind string, id string) string { return kind + helpValueMark + id }

func (form invitationForm) helpReferences() []InvitationHelp {
	var references []InvitationHelp
	if form.preaccess != "" {
		references = append(references, InvitationHelp{Kind: helpKindContact, ID: form.preaccess})
	}
	for _, value := range form.help {
		kind, id, _ := strings.Cut(value, helpValueMark)
		references = append(references, InvitationHelp{Kind: kind, ID: id})
	}
	return references
}

func (form invitationForm) request() InvitationRequest {
	return InvitationRequest{Account: form.person.Account, Output: form.output.Value(), Preaccess: form.preaccess, Help: form.helpReferences()}
}

func (form invitationForm) withGuide() invitationForm {
	context := reviewercontract.Context{Account: form.person.Account}
	for _, reference := range form.helpReferences() {
		context.GeneralHelp = append(context.GeneralHelp, reviewercontract.HelpReference{Kind: reference.Kind, ID: reference.ID})
	}
	review := reviewer.New().Review(context, form.person.Plan, form.person.Preparation)
	form.guide, form.guideProblems = review.FirstSteps, review.FirstStepsProblems
	form.guideScroll = 0
	return form
}

func (form invitationForm) preaccessChoices() []model.Choice {
	var choices []model.Choice
	for _, choice := range contactChoices(form.person.Plan.Directory) {
		if slices.Contains(form.person.PreaccessContacts, choice.Value) {
			choices = append(choices, choice)
		}
	}
	return choices
}

func (form invitationForm) helpChoices() []model.Choice {
	directory := form.person.Plan.Directory
	var choices []model.Choice
	for _, choice := range contactChoices(directory) {
		choices = append(choices, model.Choice{Value: helpValue(helpKindContact, choice.Value), Label: choice.Label})
	}
	for _, role := range directory.Roles {
		choices = append(choices, model.Choice{Value: helpValue(helpKindRole, role.ID), Label: visibleText(role.Name) + " (role)"})
	}
	return choices
}

func (form invitationForm) openedPicker() invitationForm {
	switch form.focus {
	case invitationPreaccessField:
		form.chooser = newPicker(nil, form.preaccessChoices(), single(form.preaccess), specialNone).onlyOne()
	case invitationHelpField:
		form.chooser = newPicker(nil, form.helpChoices(), form.help, specialNone)
	default:
		return form
	}
	form.choosing = true
	return form
}

func (form invitationForm) closedPicker(keep bool) invitationForm {
	form.choosing = false
	if !keep {
		return form
	}
	switch form.focus {
	case invitationPreaccessField:
		form.preaccess = first(form.chooser.picked)
	case invitationHelpField:
		form.help = slices.Clone(form.chooser.picked)
	}
	form.problems = nil
	return form.withGuide()
}

func (form invitationForm) missing() []string {
	var problems []string
	if form.preaccess == "" {
		problems = append(problems, "Pre-access help: Choose who the newcomer can reach before they have company access.")
	}
	if strings.TrimSpace(form.output.Value()) == "" {
		problems = append(problems, "File: Name the file to write, such as obk-invitation-"+form.person.Account+".zip.")
	}
	return problems
}

func (form invitationForm) preaccessName() string {
	if form.preaccess == "" {
		return notChosenText
	}
	return strings.Join(holderNames(form.person.Plan.Directory, []string{form.preaccess}), "")
}

func (form invitationForm) helpNames() string {
	if len(form.help) == 0 {
		return "nobody"
	}
	choices := form.helpChoices()
	var names []string
	for _, value := range form.help {
		if label, ok := labelOf(choices, value); ok {
			names = append(names, label)
		}
	}
	return strings.Join(names, ", ")
}
