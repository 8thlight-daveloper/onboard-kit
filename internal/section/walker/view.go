package walker

import contract "onboard-kit/contracts/section/walker"

type Comparison = contract.Comparison

const (
	ComparisonUnknown   = contract.ComparisonUnknown
	ComparisonSame      = contract.ComparisonSame
	ComparisonDifferent = contract.ComparisonDifferent
)

type View = contract.View

type SectionView = contract.SectionView

type OutcomeView = contract.OutcomeView
