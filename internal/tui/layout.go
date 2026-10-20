package tui

const (
	minimumWidth     = 60
	minimumHeight    = 12
	twoPaneWidth     = 100
	paneGap          = 3
	smallestListPane = 30
	largestListPane  = 52
	reservedRows     = 3
)

type layout struct {
	width        int
	height       int
	twoPanes     bool
	listWidth    int
	mainWidth    int
	bodyHeight   int
	documentRows int
}

func layoutFor(width int, height int) layout {
	result := layout{width: width, height: height, mainWidth: width, bodyHeight: max(height-reservedRows, 1)}
	if width >= twoPaneWidth {
		result.twoPanes = true
		result.listWidth = min(max(width/3, smallestListPane), largestListPane)
		result.mainWidth = width - result.listWidth - paneGap
	}
	result.documentRows = max(result.bodyHeight-1, 1)
	return result
}

func (l layout) tooSmall() bool {
	return l.width < minimumWidth || l.height < minimumHeight
}
