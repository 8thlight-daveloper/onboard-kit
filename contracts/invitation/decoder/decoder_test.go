package decoder_test

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash/crc32"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

type entry struct {
	Name, LocalName, Data, Compressed, Type string
	CompressedSuffix                        string
	Method, Flags, Permissions              uint16
	CreatorSystem                           *uint8
	ExternalAttributes                      *uint32
	DescriptorCompressed                    *uint32
	DescriptorUncompressed                  *uint32
	Declared                                json.RawMessage
	CRC                                     *uint32
	BadCRC, Descriptor, BadDescriptor       bool
}

type fixture struct {
	Name, Source, Code, Path string
	Entries                  []entry
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

func baseEntries() []entry {
	return []entry{{Name: "plan.obk.yaml", Data: "50"}, {Name: "records/A_1.b-2/preparation.obk.yaml", Data: "51"}, {Name: "records/A_1.b-2/record.obk.yaml", Data: "52"}, {Name: "first-steps.md", Data: "53"}}
}

func buildArchive(t *testing.T, entries []entry, metadata bool) []byte {
	t.Helper()
	var local, central bytes.Buffer
	word := func(buffer *bytes.Buffer, value any) {
		t.Helper()
		if err := binary.Write(buffer, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range entries {
		payload := unhex(t, e.Data)
		compressed := payload
		if e.Method == 8 {
			var output bytes.Buffer
			writer, err := flate.NewWriter(&output, flate.DefaultCompression)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Write(payload); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			compressed = output.Bytes()
		}
		if e.Compressed != "" {
			compressed = unhex(t, e.Compressed)
		}
		compressed = append(bytes.Clone(compressed), unhex(t, e.CompressedSuffix)...)
		declared := uint64(len(payload))
		if len(e.Declared) != 0 {
			var err error
			declared, err = strconv.ParseUint(strings.Trim(string(e.Declared), "\""), 10, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		checksum := crc32.ChecksumIEEE(payload)
		if e.CRC != nil {
			checksum = *e.CRC
		}
		if e.BadCRC {
			checksum ^= 1
		}
		flags := e.Flags
		if e.Descriptor {
			flags |= 8
		}
		localName := e.Name
		if e.LocalName != "" {
			localName = e.LocalName
		}
		var extra []byte
		declared32 := uint32(declared)
		if declared > math.MaxUint32 {
			declared32 = math.MaxUint32
			extra = binary.LittleEndian.AppendUint16(extra, 1)
			extra = binary.LittleEndian.AppendUint16(extra, 8)
			extra = binary.LittleEndian.AppendUint64(extra, declared)
		}
		if metadata {
			extra = append(extra, 0xfe, 0xca, 2, 0, 1, 2)
		}
		offset := uint32(local.Len())
		word(&local, uint32(0x04034b50))
		for _, v := range []uint16{20, flags, e.Method, 0, 0x21} {
			word(&local, v)
		}
		localCRC, localCompressed, localDeclared := checksum, uint32(len(compressed)), declared32
		if e.Descriptor {
			localCRC, localCompressed, localDeclared = 0, 0, 0
		}
		for _, v := range []uint32{localCRC, localCompressed, localDeclared} {
			word(&local, v)
		}
		word(&local, uint16(len(localName)))
		word(&local, uint16(len(extra)))
		local.WriteString(localName)
		local.Write(extra)
		local.Write(compressed)
		if e.Descriptor {
			word(&local, uint32(0x08074b50))
			descriptorCRC := checksum
			if e.BadDescriptor {
				descriptorCRC ^= 1
			}
			word(&local, descriptorCRC)
			descriptorCompressed, descriptorUncompressed := uint32(len(compressed)), declared32
			if e.DescriptorCompressed != nil {
				descriptorCompressed = *e.DescriptorCompressed
			}
			if e.DescriptorUncompressed != nil {
				descriptorUncompressed = *e.DescriptorUncompressed
			}
			word(&local, descriptorCompressed)
			word(&local, descriptorUncompressed)
		}
		word(&central, uint32(0x02014b50))
		creatorVersion := uint16(0x0314)
		if e.CreatorSystem != nil {
			creatorVersion = uint16(*e.CreatorSystem)<<8 | 20
		}
		for _, v := range []uint16{creatorVersion, 20, flags, e.Method, 0, 0x21} {
			word(&central, v)
		}
		for _, v := range []uint32{checksum, uint32(len(compressed)), declared32} {
			word(&central, v)
		}
		comment := ""
		if metadata {
			comment = "inert entry metadata"
		}
		for _, v := range []uint16{uint16(len(e.Name)), uint16(len(extra)), uint16(len(comment)), 0, 0} {
			word(&central, v)
		}
		mode := uint32(0x8000)
		switch e.Type {
		case "directory":
			mode = 0x4000
		case "symlink":
			mode = 0xa000
		case "socket":
			mode = 0xc000
		case "fifo":
			mode = 0x1000
		case "character":
			mode = 0x2000
		case "block":
			mode = 0x6000
		}
		permissions := e.Permissions
		if permissions == 0 {
			permissions = 0600
		}
		attributes := (mode | uint32(permissions)) << 16
		if e.ExternalAttributes != nil {
			attributes = *e.ExternalAttributes
		}
		word(&central, attributes)
		word(&central, offset)
		central.WriteString(e.Name)
		central.Write(extra)
		central.WriteString(comment)
	}
	start := uint32(local.Len())
	local.Write(central.Bytes())
	word(&local, uint32(0x06054b50))
	for _, v := range []uint16{0, 0, uint16(len(entries)), uint16(len(entries))} {
		word(&local, v)
	}
	word(&local, uint32(central.Len()))
	word(&local, start)
	comment := ""
	if metadata {
		comment = "inert archive metadata"
	}
	word(&local, uint16(len(comment)))
	local.WriteString(comment)
	return local.Bytes()
}

func expectedFiles(t *testing.T, entries []entry) model.InvitationFiles {
	t.Helper()
	var result model.InvitationFiles
	for _, e := range entries {
		switch {
		case e.Name == "plan.obk.yaml":
			result.Plan = unhex(t, e.Data)
		case e.Name == "first-steps.md":
			result.FirstSteps = unhex(t, e.Data)
		case strings.HasSuffix(e.Name, "/preparation.obk.yaml"):
			result.Preparation = unhex(t, e.Data)
			result.Account = strings.Split(e.Name, "/")[1]
		case strings.HasSuffix(e.Name, "/record.obk.yaml"):
			result.Record = unhex(t, e.Data)
		}
	}
	return result
}

func equalFiles(a, b model.InvitationFiles) bool {
	return a.Account == b.Account && bytes.Equal(a.Plan, b.Plan) && bytes.Equal(a.Preparation, b.Preparation) && bytes.Equal(a.Record, b.Record) && bytes.Equal(a.FirstSteps, b.FirstSteps)
}

func fileLengths(files model.InvitationFiles) [5]int {
	return [5]int{len(files.Account), len(files.Plan), len(files.Preparation), len(files.Record), len(files.FirstSteps)}
}

func check(t *testing.T, data []byte, want model.InvitationFiles, code, path string) {
	t.Helper()
	before := bytes.Clone(data)
	decoder := newDecoder()
	got, problems := decoder.Decode(data)
	if !bytes.Equal(data, before) {
		t.Fatal("input bytes changed")
	}
	if code == "" {
		if len(problems) != 0 || !equalFiles(got, want) {
			t.Fatalf("decoded payload or account mismatch; problems %#v; lengths got %v, want %v", problems, fileLengths(got), fileLengths(want))
		}
	} else {
		if !reflect.DeepEqual(got, model.InvitationFiles{}) || len(problems) != 1 {
			t.Fatalf("expected zero result and one %s error; got lengths %v, problems %#v", code, fileLengths(got), problems)
		}
		p := problems[0]
		if p.Code != code || p.Detail["path"] != path || p.Severity != model.Severity("error") || strings.TrimSpace(p.Message) == "" || p.Question != "" || p.Section != "" {
			t.Fatalf("unexpected problem %#v; want %s at %q", p, code, path)
		}
		if len(p.Detail) > 1 || (path == "" && len(p.Detail) != 0) {
			t.Fatal("unexpected diagnostic detail")
		}
		for _, marker := range []string{"private-person", "example.invalid", "A_1.b-2", "not-an-extraction-target"} {
			if strings.Contains(p.Message, marker) || strings.Contains(p.Detail["path"], marker) {
				t.Fatal("diagnostic echoes supplied personal data or raw path")
			}
		}
	}
	again, repeated := decoder.Decode(data)
	if !equalFiles(got, again) || !reflect.DeepEqual(problems, repeated) {
		t.Fatal("nondeterministic")
	}
	if code != "" && path != "" {
		problems[0].Detail["path"] = "changed"
		if repeated[0].Detail["path"] != path {
			t.Fatal("shared diagnostic map")
		}
	}
	if code == "" {
		for _, payload := range [][]byte{got.Plan, got.Preparation, got.Record, got.FirstSteps} {
			if len(payload) != 0 {
				payload[0] ^= 1
			}
		}
		if !equalFiles(again, want) || !bytes.Equal(data, before) {
			t.Fatal("result aliases another result or argument")
		}
	}
	valid := baseEntries()
	later, laterProblems := decoder.Decode(buildArchive(t, valid, false))
	if len(laterProblems) != 0 || !equalFiles(later, expectedFiles(t, valid)) {
		t.Fatal("prior call poisoned subsequent valid call")
	}
}

func TestFixtureOracle(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../encoder/testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct{ Archive string }
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 || fixtures[0].Archive == "" || !bytes.Equal(buildArchive(t, baseEntries(), false), unhex(t, fixtures[0].Archive)) {
		t.Fatal("ZIP fixture helper differs from independent complete golden")
	}
	reader, err := zip.NewReader(bytes.NewReader(buildArchive(t, baseEntries(), false)), int64(len(buildArchive(t, baseEntries(), false))))
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		got, readErr := io.ReadAll(stream)
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(got, unhex(t, baseEntries()[i].Data)) {
			t.Fatal("fixture is not an independently readable ZIP")
		}
	}
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	names := map[string]bool{}
	for _, f := range readFixtures(t) {
		if f.Name == "" || f.Source == "" || names[f.Name] || len(f.Entries) == 0 {
			t.Fatal("fixture needs unique name, source, and entries")
		}
		names[f.Name] = true
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			var want model.InvitationFiles
			if f.Code == "" {
				want = expectedFiles(t, f.Entries)
			}
			check(t, buildArchive(t, f.Entries, false), want, f.Code, f.Path)
		})
	}
}

func TestFixtureTransportControls(t *testing.T) {
	t.Parallel()
	inspect := func(t *testing.T, data []byte, entries []entry) {
		t.Helper()
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil || reader == nil {
			t.Fatalf("valid transport fixture is unreadable: %v", err)
		}
		if len(reader.File) != len(entries) {
			t.Fatal("transport fixture has wrong entry count")
		}
		for i, file := range reader.File {
			stream, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			payload, readErr := io.ReadAll(stream)
			closeErr := stream.Close()
			if readErr != nil || closeErr != nil || file.Name != entries[i].Name || !bytes.Equal(payload, unhex(t, entries[i].Data)) {
				t.Fatalf("transport fixture payload %d differs: read %v, close %v", i, readErr, closeErr)
			}
		}
	}
	for _, f := range readFixtures(t) {
		if f.Code == "" {
			t.Run(f.Name, func(t *testing.T) {
				t.Parallel()
				inspect(t, buildArchive(t, f.Entries, false), f.Entries)
			})
		}
	}
	entries := baseEntries()
	archive := buildArchive(t, entries, true)
	archive = append([]byte("inert preamble\n"), archive...)
	archive = append(archive, []byte("inert trailing bytes")...)
	inspect(t, archive, entries)
	archive = buildArchive(t, entries, false)
	archive = append(bytes.Repeat([]byte{'x'}, 2097152-len(archive)), archive...)
	inspect(t, archive, entries)
}

func TestRawMetadataFixtureOracle(t *testing.T) {
	t.Parallel()
	for _, f := range readFixtures(t) {
		if !strings.HasPrefix(f.Name, "raw-") && !strings.HasPrefix(f.Name, "nonunix-") {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			data := buildArchive(t, f.Entries, false)
			reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			for i, input := range f.Entries {
				if input.CreatorSystem != nil && reader.File[i].CreatorVersion>>8 != uint16(*input.CreatorSystem) {
					t.Fatal("fixture helper changed explicit creator system")
				}
				if input.ExternalAttributes != nil && reader.File[i].ExternalAttrs != *input.ExternalAttributes {
					t.Fatal("fixture helper changed explicit raw attributes, including zero")
				}
			}
		})
	}
}

func TestMalformedContainers(t *testing.T) {
	t.Parallel()
	archive := buildArchive(t, baseEntries(), false)
	for name, data := range map[string][]byte{"nil": nil, "empty": {}, "text": []byte("private-person@example.invalid"), "signature": []byte("PK\x03\x04"), "truncated-directory": archive[:len(archive)-10], "no-end": archive[:len(archive)-22]} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			check(t, data, model.InvitationFiles{}, "invitation.not-zip", "")
		})
	}
}

func TestHarmlessMetadata(t *testing.T) {
	t.Parallel()
	entries := baseEntries()
	archive := buildArchive(t, entries, true)
	check(t, archive, expectedFiles(t, entries), "", "")
	archive = append([]byte("inert preamble\n"), archive...)
	archive = append(archive, []byte("inert trailing bytes")...)
	check(t, archive, expectedFiles(t, entries), "", "")
}

func TestRawBudget(t *testing.T) {
	entries := baseEntries()
	archive := buildArchive(t, entries, false)
	prefix := bytes.Repeat([]byte{'x'}, 2097152-len(archive))
	data := append(prefix, archive...)
	check(t, data, expectedFiles(t, entries), "", "")
	data = append(data, 'x')
	check(t, data, model.InvitationFiles{}, "invitation.too-large", "")
	check(t, bytes.Repeat([]byte{'x'}, 2097153), model.InvitationFiles{}, "invitation.too-large", "")
}

func TestPayloadBudget(t *testing.T) {
	for _, sizes := range [][4]int{{1048576, 0, 0, 0}, {0, 1048576, 0, 0}, {0, 0, 1048576, 0}, {0, 0, 0, 1048576}, {262144, 262144, 262144, 262144}} {
		entries := baseEntries()
		for i := range entries {
			entries[i].Data = strings.Repeat("50", sizes[i])
		}
		archive := buildArchive(t, entries, false)
		check(t, archive, expectedFiles(t, entries), "", "")
		entries[3].Data += "51"
		check(t, buildArchive(t, entries, false), model.InvitationFiles{}, "invitation.too-large", "")
	}
}

func TestPathBudget(t *testing.T) {
	entries := baseEntries()
	account := strings.Repeat("a", 4067)
	entries[1].Name = "records/" + account + "/preparation.obk.yaml"
	entries[2].Name = "records/" + account + "/record.obk.yaml"
	check(t, buildArchive(t, entries, false), expectedFiles(t, entries), "", "")
	entries[1].Name = "records/" + account + "a/preparation.obk.yaml"
	entries[2].Name = "records/" + account + "a/record.obk.yaml"
	check(t, buildArchive(t, entries, false), model.InvitationFiles{}, "invitation.entry-name", "")
}

func TestActualLengthMismatch(t *testing.T) {
	entries := baseEntries()
	entries[0].Data = strings.Repeat("50", 1048577)
	entries[0].Declared = json.RawMessage("1")
	entries[0].Method = 8
	archive := buildArchive(t, entries, false)
	if len(archive) > 16384 {
		t.Fatal("bounded compressed probe unexpectedly large")
	}
	got, problems := newDecoder().Decode(archive)
	if !reflect.DeepEqual(got, model.InvitationFiles{}) || len(problems) != 1 {
		t.Fatalf("lying size produced a partial or successful result: lengths %v, problems %#v", fileLengths(got), problems)
	}
	p := problems[0]
	valid := (p.Code == "invitation.entry-data" && p.Detail["path"] == "plan") || (p.Code == "invitation.too-large" && len(p.Detail) == 0)
	if !valid || p.Severity != model.Severity("error") || strings.TrimSpace(p.Message) == "" {
		t.Fatalf("wrong bounded-read diagnostic: %#v", p)
	}
}
