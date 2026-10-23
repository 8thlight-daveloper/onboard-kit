package encoder

import "onboard-kit/internal/model"

type Encoder interface {
	Encode(model.InvitationFiles) ([]byte, []model.Problem)
}
