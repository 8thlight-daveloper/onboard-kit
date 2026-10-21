package tui

import (
	"fmt"
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

const (
	peopleListShare   = 2
	peopleListOutOf   = 3
	nameColumnShare   = 2
	nameColumnOutOf   = 5
	nameColumnMinimum = 8
	rowMarkWidth      = 2
	summarySeparator  = "  "
	stripRows         = 2
	emailColumnShare  = 2
	emailColumnOutOf  = 5
)

func (app App) peopleBody(screenFrame frame, colors palette) []string {
	switch app.people.mode {
	case peopleEditing:
		return app.formBody(screenFrame, colors)
	case peopleRemoving:
		width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
		return stackPane(pane{lines: clipLines(app.removalLines(width), screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
	}
	return app.browsingBody(screenFrame, colors)
}

func (app App) browsingBody(screenFrame frame, colors palette) []string {
	listWidth, detailWidth := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	rows, focus := app.peopleRows(listWidth, colors)
	strip := line{text: app.kindStrip(colors), raw: true}
	if screenFrame.wide() {
		shown, position := windowRows(rows, focus, screenFrame.bodyHeight-stripRows)
		listLines := append(append([]line{strip, blankLine()}, shown...), position...)
		detail := pane{lines: clipLines(app.detailLines(detailWidth), screenFrame.bodyHeight), focused: true}
		return joinPanes(pane{lines: listLines, focused: true}, detail, listWidth, detailWidth, screenFrame.bodyHeight, colors)
	}
	available := screenFrame.bodyHeight - stripRows
	shown, position := windowRows(rows, focus, max(available*peopleListShare/peopleListOutOf, minimumPickRows-1))
	content := append(append([]line{strip, blankLine()}, shown...), position...)
	if remaining := screenFrame.bodyHeight - len(content) - 1; remaining > 0 {
		content = append(append(content, blankLine()), clipLines(app.detailLines(listWidth), remaining)...)
	}
	return stackPane(pane{lines: clipLines(content, screenFrame.bodyHeight), focused: true}, listWidth, screenFrame.bodyHeight, colors)
}

func (app App) kindStrip(colors palette) string {
	var parts []string
	for _, kind := range directoryKinds {
		label := fmt.Sprintf("%s %d", kind.title(), len(entriesOf(app.plan.Directory, kind)))
		switch {
		case kind == app.people.kind && colors.styled():
			parts = append(parts, colors.render(toneHere, true, label))
		case kind == app.people.kind:
			parts = append(parts, "["+label+"]")
		default:
			parts = append(parts, colors.render(toneFaint, true, label))
		}
	}
	return strings.Join(parts, tabGap)
}

func (app App) peopleRows(width int, colors palette) ([]line, int) {
	entries := entriesOf(app.plan.Directory, app.people.kind)
	if len(entries) == 0 {
		return app.emptyListLines(width), 0
	}
	nameWidth := 0
	for _, entry := range entries {
		nameWidth = max(nameWidth, cellWidth(singleLine(entry.name)))
	}
	nameWidth = max(min(nameWidth, (width-rowMarkWidth)*nameColumnShare/nameColumnOutOf), min(nameWidth, nameColumnMinimum))
	emailWidth := app.emailColumnWidth(width)
	rows := make([]line, 0, len(entries))
	for index, entry := range entries {
		rows = append(rows, app.peopleRow(entry, nameWidth, emailWidth, index == app.people.cursor(), width, colors))
	}
	return rows, app.people.cursor()
}

func (app App) emptyListLines(width int) []line {
	noun := app.people.kind.plural()
	content := lines(toneBody, wrapText("No "+noun+" yet.", width))
	return append(content, lines(toneFaint, wrapText("Press a to add one.", width))...)
}

func (app App) peopleRow(entry directoryEntry, nameWidth int, emailWidth int, selected bool, width int, colors palette) line {
	name := truncateCells(singleLine(entry.name), nameWidth)
	text := name + spaces(nameWidth-cellWidth(name)) + summarySeparator + app.rowSummary(entry, emailWidth)
	row := hereLine(strings.TrimRight(text, " "), selected, true, width, colors)
	if !selected {
		row.tone = flagTone(app.flagsOf(app.people.kind, entry.id))
	}
	return row
}

func flagTone(flags []model.Problem) tone {
	weight := toneBody
	for _, flag := range flags {
		if flag.Severity == model.SeverityError {
			return toneProblem
		}
		weight = toneWarning
	}
	return weight
}

func (app App) rowSummary(entry directoryEntry, emailWidth int) string {
	var flagged []string
	for _, flag := range app.flagsOf(app.people.kind, entry.id) {
		flagged = append(flagged, flagWords(flag.Code))
	}
	summary := app.entrySummary(entry.id, emailWidth)
	if len(flagged) == 0 {
		return summary
	}
	return strings.TrimRight("! "+strings.Join(flagged, "; ")+summarySeparator+summary, " ")
}

func (app App) entrySummary(id string, emailWidth int) string {
	directory := app.plan.Directory
	switch app.people.kind {
	case kindRole:
		for _, role := range directory.Roles {
			if role.ID == id {
				return strings.Join(holderNames(directory, role.Holders), ", ")
			}
		}
	case kindTeam:
		for _, team := range directory.Teams {
			if team.ID == id {
				return overrideCount(len(team.Holders))
			}
		}
	default:
		for _, contact := range directory.Contacts {
			if contact.ID == id {
				email := truncateCells(singleLine(contact.Email), emailWidth)
				return strings.TrimRight(email+spaces(emailWidth-cellWidth(email))+summarySeparator+singleLine(contact.Slack), " ")
			}
		}
	}
	return ""
}

func overrideCount(count int) string {
	if count == 0 {
		return "no overrides"
	}
	return pluralWords(count, "override", "overrides")
}

func (app App) emailColumnWidth(width int) int {
	longest := 0
	for _, contact := range app.plan.Directory.Contacts {
		longest = max(longest, cellWidth(singleLine(contact.Email)))
	}
	return min(longest, width*emailColumnShare/emailColumnOutOf)
}

func (app App) detailLines(width int) []line {
	entry, ok := app.selectedEntry()
	if !ok {
		return nil
	}
	content := []line{faintLine("About this " + app.people.kind.singular()), blankLine()}
	for _, fact := range app.detailFacts(entry.id) {
		content = append(content, lines(toneFaint, wrapIndented(fact, width))...)
	}
	flags := app.flagsOf(app.people.kind, entry.id)
	if len(flags) > 0 {
		content = append(content, blankLine())
	}
	for _, flag := range flags {
		content = append(content, flagLines(flag, width)...)
	}
	return content
}

func flagLines(flag model.Problem, width int) []line {
	if flag.Severity == model.SeverityError {
		return lines(toneProblem, hanging("! problem: ", flagSentence(flag), width))
	}
	return lines(toneWarning, hanging("warning: ", flagSentence(flag), width))
}

func (app App) detailFacts(id string) []string {
	directory := app.plan.Directory
	facts := []string{"Id: " + id}
	switch app.people.kind {
	case kindRole:
		for _, role := range directory.Roles {
			if role.ID == id {
				facts = append(facts, roleFacts(directory, role)...)
			}
		}
	case kindTeam:
		for _, team := range directory.Teams {
			if team.ID == id {
				facts = append(facts, teamFacts(directory, team)...)
			}
		}
	default:
		for _, contact := range directory.Contacts {
			if contact.ID == id {
				facts = append(facts, contactFacts(directory, contact)...)
			}
		}
	}
	return facts
}

func contactFacts(directory model.Directory, contact model.Contact) []string {
	var facts []string
	for _, field := range []struct{ label, value string }{{"Email", contact.Email}, {"Slack", contact.Slack}, {"Slack id", contact.SlackID}} {
		if field.value != "" {
			facts = append(facts, field.label+": "+singleLine(field.value))
		}
	}
	var held []string
	for _, role := range directory.Roles {
		if slices.Contains(role.Holders, contact.ID) {
			held = append(held, role.Name)
		}
	}
	if len(held) > 0 {
		facts = append(facts, "Holds: "+strings.Join(held, ", "))
	}
	return facts
}

func roleFacts(directory model.Directory, role model.Role) []string {
	if len(role.Holders) == 0 {
		return []string{"Holders: nobody, the role is vacant"}
	}
	facts := []string{"Holders, in order:"}
	for index, name := range holderNames(directory, role.Holders) {
		facts = append(facts, fmt.Sprintf("  %d. %s", index+1, name))
	}
	return facts
}

func teamFacts(directory model.Directory, team model.Team) []string {
	roles := teamOverrideRoles(directory, team)
	if len(roles) == 0 {
		return []string{"No overrides. Every role uses its own holders."}
	}
	facts := []string{"Holders that differ for this team:"}
	for _, role := range roles {
		holders := team.Holders[role]
		text := "vacant for this team"
		if len(holders) > 0 {
			text = strings.Join(holderNames(directory, holders), ", ")
		}
		facts = append(facts, "  "+roleNameOf(directory, role)+": "+text)
	}
	return facts
}
