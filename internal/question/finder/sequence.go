package finder

func sequence(needs [][]int) []int {
	placed := make([]bool, len(needs))
	order := make([]int, 0, len(needs))
	for len(order) < len(needs) {
		next := firstReady(needs, placed)
		placed[next] = true
		order = append(order, next)
	}
	return order
}

func firstReady(needs [][]int, placed []bool) int {
	firstUnplaced := -1
	for position := range needs {
		if placed[position] {
			continue
		}
		if firstUnplaced < 0 {
			firstUnplaced = position
		}
		if allPlaced(needs[position], placed) {
			return position
		}
	}
	return firstUnplaced
}

func allPlaced(needs []int, placed []bool) bool {
	for _, need := range needs {
		if !placed[need] {
			return false
		}
	}
	return true
}
