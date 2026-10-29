package store

import (
	"reflect"

	contract "onboard-kit/contracts/preparation/store"
	"onboard-kit/internal/model"
	preparationdecoder "onboard-kit/internal/preparation/decoder"
)

func (store) Save(request contract.SaveRequest) contract.SaveResult {
	state, ok := request.Loaded.(*loadedState)
	if !ok || state == nil {
		return contract.SaveResult{Code: contract.CodeLoadedState}
	}
	if len(request.Reviewed) > byteLimit {
		return contract.SaveResult{Code: contract.CodeTooLarge}
	}
	draft, ok := decodeDraft(state, request.Reviewed)
	if !ok {
		return contract.SaveResult{Code: contract.CodeDraft}
	}
	if !unchangedSinceLoad(state) {
		return contract.SaveResult{Code: contract.CodeStale}
	}
	if state.preparation.exists && reflect.DeepEqual(draft, state.Preparation()) {
		return contract.SaveResult{Code: contract.CodeUnchanged}
	}
	if state.directoriesMissing && !request.CreateDirectories {
		return contract.SaveResult{Code: contract.CodeDirectories}
	}
	return publish(state, request.Reviewed)
}

func decodeDraft(state *loadedState, reviewed []byte) (model.Preparation, bool) {
	if len(reviewed) == 0 {
		return model.Preparation{}, false
	}
	value, problems := preparationdecoder.New().Decode(reviewed)
	if len(problems) != 0 || value.Account != state.account || value.PlanID != state.selection.planValue.ID {
		return model.Preparation{}, false
	}
	return value, true
}
