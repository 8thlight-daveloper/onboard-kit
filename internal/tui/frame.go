package tui

const (
	minimumWidth  = 60
	minimumHeight = 12
	wideWidth     = 120
	sideMargin    = 2
	columnGap     = 3
	chromeRows    = 5
	readingWidth  = 100
	narrowestPane = 24
)

type frame struct {
	width      int
	height     int
	bodyHeight int
}

func newFrame(width int, height int) frame {
	return frame{width: width, height: height, bodyHeight: max(height-chromeRows, 1)}
}

func (screen frame) tooSmall() bool {
	return screen.width < minimumWidth || screen.height < minimumHeight
}

func (screen frame) wide() bool { return screen.width >= wideWidth }

func (screen frame) innerWidth() int { return max(screen.width-2*sideMargin, 1) }

func (screen frame) columns(share int, outOf int) (int, int) {
	inner := screen.innerWidth()
	if !screen.wide() {
		return min(inner, readingWidth), 0
	}
	first := max(inner*share/outOf, narrowestPane)
	return first, max(inner-first-columnGap, narrowestPane)
}

func (screen frame) documentRows() int { return max(screen.bodyHeight-1, 1) }
