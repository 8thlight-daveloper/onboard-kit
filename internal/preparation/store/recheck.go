package store

import "golang.org/x/sys/unix"

func openUnchangedRoot(state *loadedState) (int, bool) {
	root, err := openDirectory(state.selection.rootPath)
	if err != nil {
		return -1, false
	}
	var stat unix.Stat_t
	if unix.Fstat(root, &stat) != nil || identityOf(stat) != state.selection.rootIdentity {
		unix.Close(root)
		return -1, false
	}
	return root, true
}

func unchangedSinceLoad(state *loadedState) bool {
	root, ok := openUnchangedRoot(state)
	if !ok {
		return false
	}
	defer unix.Close(root)
	plan, err := readPlanFile(state.selection.planPath)
	if err != nil || !sameSnapshot(plan, state.selection.plan) {
		return false
	}
	current, _, code := readDestination(root, state.destination, plan.facts.identity)
	return code == "" && sameSnapshot(current, state.preparation)
}
