package decoder

import (
	"archive/zip"
	"bytes"
	"errors"

	"onboard-kit/internal/model"
)

const (
	archiveLimit = 2 * 1024 * 1024
	payloadLimit = 1024 * 1024
	nameLimit    = 4096
	roleCount    = 4
)

type Decoder struct{}

func New() Decoder {
	return Decoder{}
}

func (Decoder) Decode(data []byte) (model.InvitationFiles, []model.Problem) {
	if len(data) > archiveLimit {
		return failure("too-large", "Reduce the invitation archive to the supported size.", "")
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return failure("not-zip", "Obtain a complete, readable invitation ZIP archive.", "")
	}
	if len(archive.File) != roleCount {
		return failure("entries", "Recreate the invitation with exactly four required files.", "")
	}
	entries, account, valid := identifyEntries(archive.File)
	if !valid {
		return failure("entry-name", "Recreate the invitation with unique supported names for one account.", "")
	}
	if problems := validateTypes(entries); len(problems) != 0 {
		return model.InvitationFiles{}, problems
	}
	if problems := validateEncodings(entries); len(problems) != 0 {
		return model.InvitationFiles{}, problems
	}
	if !withinDeclaredBudget(entries) {
		return failure("too-large", "Reduce the combined invitation file contents to the supported size.", "")
	}
	payloads, problems := readPayloads(entries)
	if len(problems) != 0 {
		return model.InvitationFiles{}, problems
	}
	return model.InvitationFiles{
		Account: account, Plan: payloads[0], Preparation: payloads[1],
		Record: payloads[2], FirstSteps: payloads[3],
	}, nil
}

func failure(code, message, path string) (model.InvitationFiles, []model.Problem) {
	problem := model.Problem{
		Code: "invitation." + code, Severity: model.SeverityError, Message: message,
	}
	if path != "" {
		problem.Detail = map[string]string{"path": path}
	}
	return model.InvitationFiles{}, []model.Problem{problem}
}
