package decoder

import (
	"hash/crc32"
	"io"

	"onboard-kit/internal/model"
)

func readPayloads(entries []entry) ([roleCount][]byte, []model.Problem) {
	var ordered [roleCount]entry
	for _, entry := range entries {
		ordered[entry.role] = entry
	}
	var payloads [roleCount][]byte
	remaining := int64(payloadLimit)
	for index, entry := range ordered {
		payload, problems := readPayload(entry, remaining)
		if len(problems) != 0 {
			return [roleCount][]byte{}, problems
		}
		remaining -= int64(len(payload))
		payloads[index] = payload
	}
	return payloads, nil
}

func readPayload(entry entry, remaining int64) ([]byte, []model.Problem) {
	reader, err := entry.file.Open()
	if err != nil {
		return nil, dataProblem(entry.path)
	}
	payload, readErr := io.ReadAll(io.LimitReader(reader, remaining+1))
	closeErr := reader.Close()
	if int64(len(payload)) > remaining {
		_, problems := failure("too-large", "Reduce the combined invitation file contents to the supported size.", "")
		return nil, problems
	}
	if readErr != nil || closeErr != nil || uint64(len(payload)) != entry.file.UncompressedSize64 ||
		crc32.ChecksumIEEE(payload) != entry.file.CRC32 {
		return nil, dataProblem(entry.path)
	}
	return payload, nil
}

func dataProblem(path string) []model.Problem {
	_, problems := failure("entry-data", "Obtain a complete invitation with intact file contents and checksums.", path)
	return problems
}
