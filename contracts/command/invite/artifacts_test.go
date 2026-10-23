package invite_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

type change struct {
	Old string
	New string
}

type exportCase struct {
	Name               string
	Source             string
	PlanPrefix         string
	PlanAppend         string
	PlanChanges        []change
	PreparationAppend  string
	PreparationChanges []change
	Helper             string
	Extra              []string
	Exit               int
	Code               string
	Baseline           bool
	Problems           []projectedProblem
	Blockers           []string
}

func applyChanges(t *testing.T, source []byte, changes []change) []byte {
	t.Helper()
	text := string(source)
	for _, item := range changes {
		if strings.Count(text, item.Old) != 1 {
			t.Fatalf("literal fixture change has no unique source: %q", item.Old)
		}
		text = strings.Replace(text, item.Old, item.New, 1)
	}
	return []byte(text)
}

func exportedPayloads(t *testing.T, path string) map[string][]byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatal("published archive must be a private regular file")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []string{"plan.obk.yaml", "records/newcomer/preparation.obk.yaml", "records/newcomer/record.obk.yaml", "first-steps.md"}
	if reader.Comment != "" || len(reader.File) != len(wantNames) || len(data) > 2097152 {
		t.Fatal("archive comment, entry count or encoded budget differs")
	}
	if len(data) < 22 || !bytes.Equal(data[:4], []byte{'P', 'K', 3, 4}) || !bytes.Equal(data[len(data)-22:len(data)-18], []byte{'P', 'K', 5, 6}) {
		t.Fatal("archive has preamble, trailing material or unexpected EOCD profile")
	}
	payloads := make(map[string][]byte)
	total := 0
	local := 0
	for index, file := range reader.File {
		if file.Name != wantNames[index] || file.Method != zip.Store || file.Flags != 0 || file.CreatorVersion != 0x0314 || file.ReaderVersion != 20 || file.ModifiedTime != 0 || file.ModifiedDate != 0x21 || file.ExternalAttrs != 0x81800000 || file.Mode().Perm() != 0o600 || !file.Mode().IsRegular() || len(file.Extra) != 0 || file.Comment != "" {
			t.Fatalf("noncanonical entry metadata for index %d: %#v", index, file.FileHeader)
		}
		body, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		payload, readErr := io.ReadAll(io.LimitReader(body, 1048577))
		closeErr := body.Close()
		if readErr != nil || closeErr != nil || len(payload) > 1048576 || uint64(len(payload)) != file.UncompressedSize64 || file.CompressedSize64 != file.UncompressedSize64 || crc32.ChecksumIEEE(payload) != file.CRC32 {
			t.Fatal("payload length/checksum differs")
		}
		if len(data)-local < 30 || binary.LittleEndian.Uint32(data[local:]) != 0x04034b50 || binary.LittleEndian.Uint16(data[local+26:]) != uint16(len(file.Name)) || binary.LittleEndian.Uint16(data[local+28:]) != 0 {
			t.Fatal("local entry routing/extra data differs")
		}
		if binary.LittleEndian.Uint16(data[local+4:]) != 20 || binary.LittleEndian.Uint16(data[local+6:]) != 0 || binary.LittleEndian.Uint16(data[local+8:]) != 0 || binary.LittleEndian.Uint16(data[local+10:]) != 0 || binary.LittleEndian.Uint16(data[local+12:]) != 0x21 || binary.LittleEndian.Uint32(data[local+14:]) != crc32.ChecksumIEEE(payload) || binary.LittleEndian.Uint32(data[local+18:]) != uint32(len(payload)) || binary.LittleEndian.Uint32(data[local+22:]) != uint32(len(payload)) {
			t.Fatal("local header profile differs from independent fixed metadata and payload CRC")
		}
		start := local + 30 + len(file.Name)
		if start+len(payload) > len(data) || string(data[local+30:start]) != file.Name || !bytes.Equal(data[start:start+len(payload)], payload) {
			t.Fatal("local entry bytes differ from canonical routing and literal payload")
		}
		local = start + len(payload)
		total += len(payload)
		payloads[file.Name] = payload
	}
	if total > 1048576 || len(data)-local < 4 || binary.LittleEndian.Uint32(data[local:]) != 0x02014b50 {
		t.Fatal("aggregate payload budget or central directory placement differs")
	}
	centralSize := 0
	for _, name := range wantNames {
		centralSize += 46 + len(name)
	}
	end := data[len(data)-22:]
	if len(data) != local+centralSize+22 || binary.LittleEndian.Uint16(end[4:]) != 0 || binary.LittleEndian.Uint16(end[6:]) != 0 || binary.LittleEndian.Uint16(end[8:]) != 4 || binary.LittleEndian.Uint16(end[10:]) != 4 || binary.LittleEndian.Uint32(end[12:]) != uint32(centralSize) || binary.LittleEndian.Uint32(end[16:]) != uint32(local) || binary.LittleEndian.Uint16(end[20:]) != 0 {
		t.Fatal("EOCD profile, exact offsets or extra archive material differs")
	}
	return payloads
}

func checkSeed(t *testing.T, payloads map[string][]byte) {
	t.Helper()
	hash := sha256.Sum256(payloads["plan.obk.yaml"])
	heldHash := hex.EncodeToString(hash[:])
	if heldHash[0] >= '0' && heldHash[0] <= '9' {
		heldHash = "\"" + heldHash + "\""
	}
	want := strings.Replace(string(literal(t, "record.yaml")), baselineHash, heldHash, 1)
	if string(payloads["records/newcomer/record.obk.yaml"]) != want {
		t.Fatal("seed must be exactly identity/reference plus hash of canonical embedded plan")
	}
}

func TestLiteralExportCases(t *testing.T) {
	var cases []exportCase
	decoder := json.NewDecoder(bytes.NewReader(literal(t, "cases.json")))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 18 {
		t.Fatal("expected eighteen independently sourced export cases")
	}
	for _, item := range cases {
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()
			if item.Name == "" || item.Source == "" {
				t.Fatal("missing case provenance")
			}
			s := newSandbox(t)
			plan := append([]byte(item.PlanPrefix), applyChanges(t, literal(t, "plan.yaml"), item.PlanChanges)...)
			plan = append(plan, []byte(item.PlanAppend)...)
			preparation := append(applyChanges(t, literal(t, "preparation.yaml"), item.PreparationChanges), []byte(item.PreparationAppend)...)
			writeLiteral(t, s.plan, plan)
			writeLiteral(t, s.preparation, preparation)
			args := s.args()
			if item.Helper != "" {
				args[len(args)-1] = item.Helper
			}
			args = append(args, item.Extra...)
			before := tree(t, s.root)
			got := run(newCommand(), args)
			readReport(t, got, item.Code, item.Exit, item.Problems, item.Blockers)
			preserved(t, s, before, item.Exit == 0)
			if item.Exit != 0 {
				return
			}
			payloads := exportedPayloads(t, s.out)
			checkSeed(t, payloads)
			if !bytes.Equal(payloads["first-steps.md"], literal(t, "first-steps.md")) {
				t.Fatal("literal normative guide changed, copied personal data, or invented instructions")
			}
			if item.Baseline {
				for name, fixture := range map[string]string{"plan.obk.yaml": "plan.yaml", "records/newcomer/preparation.obk.yaml": "preparation.yaml", "records/newcomer/record.obk.yaml": "record.yaml"} {
					if !bytes.Equal(payloads[name], literal(t, fixture)) {
						t.Fatalf("literal canonical payload differs: %s", name)
					}
				}
				if baseline := sha256.Sum256(literal(t, "plan.yaml")); hex.EncodeToString(baseline[:]) != baselineHash {
					t.Fatal("independently pinned baseline hash is wrong")
				}
			}
		})
	}
}

func TestRepeatedIndependentExportsAreByteIdentical(t *testing.T) {
	shared := newCommand()
	var previous []byte
	for range 2 {
		s := newSandbox(t)
		before := tree(t, s.root)
		readReport(t, run(shared, s.args()), "invite.written", 0, nil, nil)
		preserved(t, s, before, true)
		data, err := os.ReadFile(s.out)
		if err != nil {
			t.Fatal(err)
		}
		if previous != nil && !reflect.DeepEqual(previous, data) {
			t.Fatal("identical sources produced different archive bytes across paths or calls")
		}
		previous = data
	}
}
