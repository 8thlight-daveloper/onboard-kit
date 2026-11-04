package tui

type screen int

const (
	screenWalk screen = iota
	screenOverview
	screenReview
	screenSections
	screenContacts
	screenPreparation
	screenInvitation
)
