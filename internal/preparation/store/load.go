package store

import (
	contract "onboard-kit/contracts/preparation/store"
	preparationdecoder "onboard-kit/internal/preparation/decoder"

	"golang.org/x/sys/unix"
)

func (store) Load(request contract.LoadRequest) contract.LoadResult {
	root, selected, code := openSelection(request.Root, request.Plan)
	if code != "" {
		return contract.LoadResult{Code: code}
	}
	defer unix.Close(root)
	destination, ok := locate(selected.planValue.ResultsPath, request.Account)
	if !ok {
		return contract.LoadResult{Code: contract.CodePath}
	}
	held, missing, code := readDestination(root, destination, selected.plan.facts.identity)
	if code != "" {
		return contract.LoadResult{Code: code}
	}
	if held.exists {
		value, problems := preparationdecoder.New().Decode(held.content)
		if len(problems) != 0 {
			return contract.LoadResult{Code: contract.CodeSource}
		}
		if value.Account != request.Account || value.PlanID != selected.planValue.ID {
			return contract.LoadResult{Code: contract.CodeIdentity}
		}
	}
	return contract.LoadResult{Code: contract.CodeLoaded, Loaded: &loadedState{selection: selected, account: request.Account, destination: destination, preparation: held, directoriesMissing: missing}}
}
