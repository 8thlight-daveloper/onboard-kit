package decoder

import "onboard-kit/internal/model"

type Decoder interface {
	Decode([]byte) (model.InvitationFiles, []model.Problem)
}
