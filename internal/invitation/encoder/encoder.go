package encoder

import (
	"encoding/binary"
	"hash/crc32"

	contract "onboard-kit/contracts/invitation/encoder"
	"onboard-kit/internal/model"
)

type encoder struct{}

func New() contract.Encoder {
	return encoder{}
}

const (
	pathLimit         = 4096
	payloadLimit      = 1048576
	preparationPrefix = "records/"
	preparationSuffix = "/preparation.obk.yaml"
)

type entry struct {
	name     string
	payload  []byte
	checksum uint32
	offset   uint32
}

func (encoder) Encode(files model.InvitationFiles) ([]byte, []model.Problem) {
	if !validAccount(files.Account) {
		return nil, problem("invitation.account", "Account must start with an ASCII letter or digit and contain only ASCII letters, digits, dots, underscores, or hyphens; supply a matching account.", "account")
	}
	if len(files.Account) > pathLimit-len(preparationPrefix)-len(preparationSuffix) {
		return nil, problem("invitation.path-too-long", "The generated preparation path must fit 4096 bytes; supply a shorter account.", "account")
	}
	payloads := [][]byte{files.Plan, files.Preparation, files.Record, files.FirstSteps}
	total := 0
	for _, payload := range payloads {
		if len(payload) > payloadLimit-total {
			return nil, problem("invitation.too-large", "The combined payloads must fit 1048576 bytes; reduce their total size.", "")
		}
		total += len(payload)
	}
	entries := []entry{
		{name: "plan.obk.yaml", payload: files.Plan},
		{name: preparationPrefix + files.Account + preparationSuffix, payload: files.Preparation},
		{name: preparationPrefix + files.Account + "/record.obk.yaml", payload: files.Record},
		{name: "first-steps.md", payload: files.FirstSteps},
	}
	archiveSize := total + 22
	for _, item := range entries {
		archiveSize += 30 + 46 + 2*len(item.name)
	}
	archive := make([]byte, 0, archiveSize)
	for index := range entries {
		item := &entries[index]
		item.offset = uint32(len(archive))
		item.checksum = crc32.ChecksumIEEE(item.payload)
		archive = appendLocal(archive, *item)
	}
	centralOffset := uint32(len(archive))
	for _, item := range entries {
		archive = appendCentral(archive, item)
	}
	centralSize := uint32(len(archive)) - centralOffset
	archive = binary.LittleEndian.AppendUint32(archive, 0x06054b50)
	archive = appendWords(archive, 0, 0, 4, 4)
	archive = binary.LittleEndian.AppendUint32(archive, centralSize)
	archive = binary.LittleEndian.AppendUint32(archive, centralOffset)
	archive = appendWords(archive, 0)
	return archive, nil
}

func validAccount(account string) bool {
	if len(account) == 0 || !asciiAlphanumeric(account[0]) {
		return false
	}
	for index := 1; index < len(account); index++ {
		character := account[index]
		if !asciiAlphanumeric(character) && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphanumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9'
}

func problem(code, message, path string) []model.Problem {
	result := model.Problem{Code: code, Severity: model.SeverityError, Message: message}
	if path != "" {
		result.Detail = map[string]string{"path": path}
	}
	return []model.Problem{result}
}

func appendLocal(archive []byte, item entry) []byte {
	archive = binary.LittleEndian.AppendUint32(archive, 0x04034b50)
	archive = appendWords(archive, 20, 0, 0, 0, 0x0021)
	archive = appendPayloadMetadata(archive, item)
	archive = appendWords(archive, uint16(len(item.name)), 0)
	archive = append(archive, item.name...)
	return append(archive, item.payload...)
}

func appendCentral(archive []byte, item entry) []byte {
	archive = binary.LittleEndian.AppendUint32(archive, 0x02014b50)
	archive = appendWords(archive, 0x0314, 20, 0, 0, 0, 0x0021)
	archive = appendPayloadMetadata(archive, item)
	archive = appendWords(archive, uint16(len(item.name)), 0, 0, 0, 0)
	archive = binary.LittleEndian.AppendUint32(archive, 0x81800000)
	archive = binary.LittleEndian.AppendUint32(archive, item.offset)
	return append(archive, item.name...)
}

func appendPayloadMetadata(archive []byte, item entry) []byte {
	archive = binary.LittleEndian.AppendUint32(archive, item.checksum)
	archive = binary.LittleEndian.AppendUint32(archive, uint32(len(item.payload)))
	return binary.LittleEndian.AppendUint32(archive, uint32(len(item.payload)))
}

func appendWords(archive []byte, words ...uint16) []byte {
	for _, word := range words {
		archive = binary.LittleEndian.AppendUint16(archive, word)
	}
	return archive
}
