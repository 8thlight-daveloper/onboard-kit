package encoder_test

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

type fixture struct {
	Name, Source, Account, Plan, Preparation, Record, FirstSteps, Archive string
}

func readFixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []fixture
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty fixture manifest")
	}
	return cases
}

func unhex(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func (f fixture) input(t *testing.T) model.InvitationFiles {
	t.Helper()
	return model.InvitationFiles{Account: f.Account, Plan: unhex(t, f.Plan), Preparation: unhex(t, f.Preparation), Record: unhex(t, f.Record), FirstSteps: unhex(t, f.FirstSteps)}
}

func canonicalArchive(input model.InvitationFiles) []byte {
	names := []string{"plan.obk.yaml", "records/" + input.Account + "/preparation.obk.yaml", "records/" + input.Account + "/record.obk.yaml", "first-steps.md"}
	payloads := [][]byte{input.Plan, input.Preparation, input.Record, input.FirstSteps}
	var local, central bytes.Buffer
	word := func(buffer *bytes.Buffer, value any) {
		if err := binary.Write(buffer, binary.LittleEndian, value); err != nil {
			panic(err)
		}
	}
	for i, name := range names {
		offset := uint32(local.Len())
		checksum, size := crc32.ChecksumIEEE(payloads[i]), uint32(len(payloads[i]))
		word(&local, uint32(0x04034b50))
		for _, v := range []uint16{20, 0, 0, 0, 0x21} {
			word(&local, v)
		}
		for _, v := range []uint32{checksum, size, size} {
			word(&local, v)
		}
		word(&local, uint16(len(name)))
		word(&local, uint16(0))
		local.WriteString(name)
		local.Write(payloads[i])
		word(&central, uint32(0x02014b50))
		for _, v := range []uint16{0x0314, 20, 0, 0, 0, 0x21} {
			word(&central, v)
		}
		for _, v := range []uint32{checksum, size, size} {
			word(&central, v)
		}
		for _, v := range []uint16{uint16(len(name)), 0, 0, 0, 0} {
			word(&central, v)
		}
		word(&central, uint32(0x81800000))
		word(&central, offset)
		central.WriteString(name)
	}
	start := uint32(local.Len())
	local.Write(central.Bytes())
	word(&local, uint32(0x06054b50))
	for _, v := range []uint16{0, 0, 4, 4} {
		word(&local, v)
	}
	word(&local, uint32(central.Len()))
	word(&local, start)
	word(&local, uint16(0))
	return local.Bytes()
}

func TestFixtureOracle(t *testing.T) {
	t.Parallel()
	f := readFixtures(t)[0]
	if !bytes.Equal(canonicalArchive(f.input(t)), unhex(t, f.Archive)) {
		t.Fatal("record-layout oracle differs from pinned complete ZIP golden")
	}
	if crc32.ChecksumIEEE([]byte("P")) != 0xb969be79 || crc32.ChecksumIEEE([]byte("Q")) != 0xce6e8eef || crc32.ChecksumIEEE([]byte("R")) != 0x5767df55 || crc32.ChecksumIEEE([]byte("S")) != 0x2060efc3 {
		t.Fatal("fixture CRC constants are incorrect")
	}
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	names := map[string]bool{}
	for _, f := range readFixtures(t) {
		if f.Name == "" || f.Source == "" || names[f.Name] {
			t.Fatal("fixture requires unique name and source")
		}
		names[f.Name] = true
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			input := f.input(t)
			before := f.input(t)
			encoder := newEncoder()
			got, problems := encoder.Encode(input)
			want := canonicalArchive(before)
			if !bytes.Equal(got, want) || len(problems) != 0 {
				t.Fatalf("archive mismatch: got %d bytes, want %d; problems %#v", len(got), len(want), problems)
			}
			if !reflect.DeepEqual(input, before) {
				t.Fatal("input changed")
			}
			again, repeated := encoder.Encode(input)
			if !bytes.Equal(got, again) || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("nondeterministic")
			}
			got[0] ^= 1
			if !bytes.Equal(again, want) {
				t.Fatal("separate output aliases prior output")
			}
			inspectArchive(t, again, before)
		})
	}
}

func inspectArchive(t *testing.T, data []byte, input model.InvitationFiles) {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.File) != 4 || reader.Comment != "" {
		t.Fatal("incorrect entry count or archive comment")
	}
	want := [][]byte{input.Plan, input.Preparation, input.Record, input.FirstSteps}
	for i, file := range reader.File {
		if file.Method != zip.Store || file.Mode() != 0600 || file.Modified.Year() != 1980 || len(file.Extra) != 0 || file.Comment != "" {
			t.Fatalf("unexpected metadata at role %d", i)
		}
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		payload, readErr := io.ReadAll(stream)
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(payload, want[i]) {
			t.Fatalf("payload mismatch at role %d", i)
		}
	}
}

func assertProblem(t *testing.T, input model.InvitationFiles, code, path string) {
	t.Helper()
	encoder := newEncoder()
	data, problems := encoder.Encode(input)
	if len(data) != 0 || len(problems) != 1 {
		t.Fatalf("expected no bytes and one %s problem, got %d bytes and %#v", code, len(data), problems)
	}
	p := problems[0]
	if p.Code != code || p.Severity != model.Severity("error") || strings.TrimSpace(p.Message) == "" || p.Question != "" || p.Section != "" || p.Detail["path"] != path {
		t.Fatalf("unexpected problem %#v; want %s at %q", p, code, path)
	}
	if len(p.Detail) > 1 || (path == "" && len(p.Detail) != 0) {
		t.Fatal("unexpected diagnostic detail")
	}
	_, repeated := encoder.Encode(input)
	if !reflect.DeepEqual(problems, repeated) {
		t.Fatal("nondeterministic diagnostic")
	}
	if path != "" {
		problems[0].Detail["path"] = "changed"
		if repeated[0].Detail["path"] != path {
			t.Fatal("shared diagnostic map")
		}
	}
	valid := model.InvitationFiles{Account: "a"}
	got, later := encoder.Encode(valid)
	if !bytes.Equal(got, canonicalArchive(valid)) || len(later) != 0 {
		t.Fatal("failed call poisoned next valid call")
	}
}

func TestInvalidAccounts(t *testing.T) {
	t.Parallel()
	for _, account := range []string{"", ".", "..", "-a", "_a", " a", "a ", "a/b", "a\\b", "a:b", "a%2fb", "a\x00b", "a\nb", "é", string([]byte{0xff})} {
		t.Run(hex.EncodeToString([]byte(account)), func(t *testing.T) {
			t.Parallel()
			assertProblem(t, model.InvitationFiles{Account: account}, "invitation.account", "account")
		})
	}
}

func TestPayloadBudget(t *testing.T) {
	for _, sizes := range [][4]int{{1048576, 0, 0, 0}, {0, 1048576, 0, 0}, {0, 0, 1048576, 0}, {0, 0, 0, 1048576}, {262144, 262144, 262144, 262144}} {
		input := model.InvitationFiles{Account: "a", Plan: bytes.Repeat([]byte{1}, sizes[0]), Preparation: bytes.Repeat([]byte{2}, sizes[1]), Record: bytes.Repeat([]byte{3}, sizes[2]), FirstSteps: bytes.Repeat([]byte{4}, sizes[3])}
		got, problems := newEncoder().Encode(input)
		if !bytes.Equal(got, canonicalArchive(input)) || len(problems) != 0 || len(got) > 2097152 {
			t.Fatalf("inclusive budget failed for %v: %d bytes, %#v", sizes, len(got), problems)
		}
		input.FirstSteps = append(input.FirstSteps, 5)
		assertProblem(t, input, "invitation.too-large", "")
	}
}

func TestPathBudget(t *testing.T) {
	input := model.InvitationFiles{Account: strings.Repeat("a", 4067)}
	got, problems := newEncoder().Encode(input)
	if !bytes.Equal(got, canonicalArchive(input)) || len(problems) != 0 {
		t.Fatalf("inclusive 4096-byte path boundary refused: %#v", problems)
	}
	input.Account += "a"
	assertProblem(t, input, "invitation.path-too-long", "account")
}

func TestPrecedenceAndPrivacy(t *testing.T) {
	marker := "private-person@example.invalid"
	input := model.InvitationFiles{Account: marker + strings.Repeat("a", 4068), Plan: bytes.Repeat([]byte{0}, 1048577)}
	assertProblem(t, input, "invitation.account", "account")
	_, problems := newEncoder().Encode(input)
	if strings.Contains(problems[0].Message, marker) {
		t.Fatal("diagnostic leaked supplied account")
	}
	input.Account = strings.Repeat("a", 4068)
	assertProblem(t, input, "invitation.path-too-long", "account")
}
