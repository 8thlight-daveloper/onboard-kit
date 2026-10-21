package tui

import (
	"fmt"
	"slices"

	"onboard-kit/internal/model"
)

const (
	specialNone    = -1
	specialInherit = 0
	specialVacant  = 1
)

type picker struct {
	leading  []string
	contacts []model.Choice
	cursor   int
	special  int
	picked   []string
	single   bool
}

func newPicker(leading []string, contacts []model.Choice, picked []string, special int) picker {
	chooser := picker{leading: leading, contacts: contacts, special: special}
	for _, id := range picked {
		if _, ok := labelOf(contacts, id); ok && !slices.Contains(chooser.picked, id) {
			chooser.picked = append(chooser.picked, id)
		}
	}
	switch {
	case special >= 0:
		chooser.cursor = special
	case len(chooser.picked) > 0:
		chooser.cursor = len(leading) + slices.IndexFunc(contacts, func(each model.Choice) bool { return each.Value == chooser.picked[0] })
	}
	return chooser
}

func (chooser picker) onlyOne() picker {
	chooser.single = true
	return chooser
}

func (chooser picker) rows() int { return len(chooser.leading) + len(chooser.contacts) }

func (chooser picker) moved(delta int) picker {
	chooser.cursor = max(min(chooser.cursor+delta, chooser.rows()-1), 0)
	return chooser
}

func (chooser picker) toggled() picker {
	if chooser.rows() == 0 {
		return chooser
	}
	if chooser.cursor < len(chooser.leading) {
		chooser.special = chooser.cursor
		chooser.picked = nil
		return chooser
	}
	id := chooser.contacts[chooser.cursor-len(chooser.leading)].Value
	if chooser.single {
		chooser.picked = []string{id}
		chooser.special = specialNone
		return chooser
	}
	if index := slices.Index(chooser.picked, id); index >= 0 {
		chooser.picked = slices.Delete(slices.Clone(chooser.picked), index, index+1)
	} else {
		chooser.picked = append(slices.Clone(chooser.picked), id)
	}
	switch {
	case len(chooser.picked) > 0:
		chooser.special = specialNone
	case len(chooser.leading) > 0:
		chooser.special = specialInherit
	}
	return chooser
}

func (chooser picker) rowText(row int) string {
	if row < len(chooser.leading) {
		mark := "( ) "
		if chooser.special == row {
			mark = "(o) "
		}
		return mark + chooser.leading[row]
	}
	choice := chooser.contacts[row-len(chooser.leading)]
	label, _ := labelOf(chooser.contacts, choice.Value)
	if chooser.single {
		if slices.Contains(chooser.picked, choice.Value) {
			return "(o) " + label
		}
		return "( ) " + label
	}
	if order := slices.Index(chooser.picked, choice.Value) + 1; order > 0 {
		return fmt.Sprintf("[%d] %s", order, label)
	}
	return "[ ] " + label
}

func (chooser picker) orderText() string {
	if len(chooser.picked) < 2 {
		return ""
	}
	var names []string
	for _, id := range chooser.picked {
		label, _ := labelOf(chooser.contacts, id)
		names = append(names, label)
	}
	return "Order: " + joinThen(names)
}

func joinThen(names []string) string {
	text := ""
	for index, name := range names {
		if index > 0 {
			text += ", then "
		}
		text += name
	}
	return text
}
