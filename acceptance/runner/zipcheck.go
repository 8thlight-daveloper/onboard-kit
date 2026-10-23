package runner

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

const zipArchiveLimit int64 = 2 << 20
const zipPayloadLimit int64 = 1 << 20

func checkZIPFile(path string, expectation zipExpectation, folder, repositoryRoot string) []Failure {
	label := "file " + path
	target, err := resolve(folder, path)
	if err != nil {
		return []Failure{{Expectation: label + " zip", Message: err.Error()}}
	}
	content, err := readZIPRegular(target, zipArchiveLimit)
	if os.IsNotExist(err) {
		return []Failure{{Expectation: label, Message: "the file does not exist"}}
	}
	if err != nil {
		return []Failure{{Expectation: label + " zip", Message: err.Error()}}
	}
	if err := checkZIP(content, expectation, repositoryRoot); err != nil {
		return []Failure{{Expectation: label + " zip", Message: err.Error()}}
	}
	return nil
}

func readZIPRegular(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("expected a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err = file.Stat()
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("expected a regular file")
	}
	if err != nil {
		file.Close()
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds %d byte limit", limit)
	}
	return data, nil
}

func checkZIP(content []byte, expectation zipExpectation, repositoryRoot string) error {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil && (!errors.Is(err, zip.ErrInsecurePath) || archive == nil) {
		return fmt.Errorf("cannot inspect ZIP: %w", err)
	}
	if len(archive.File) != len(expectation.entries) {
		return fmt.Errorf("ZIP entry count: expected %d, got %d", len(expectation.entries), len(archive.File))
	}
	remaining := zipPayloadLimit
	for index, file := range archive.File {
		wanted := expectation.entries[index]
		if err := zipMetadata(file, wanted); err != nil {
			return fmt.Errorf("entry %d %q: %w", index+1, file.Name, err)
		}
		payload, err := zipPayload(file, remaining)
		if err != nil {
			return fmt.Errorf("entry %d %q payload: %w", index+1, file.Name, err)
		}
		remaining -= int64(len(payload))
		fixture, err := zipFixture(repositoryRoot, wanted.sameAs, int64(len(payload)))
		if err != nil {
			return fmt.Errorf("entry %d %q same-as %q: %w", index+1, file.Name, wanted.sameAs, err)
		}
		if !bytes.Equal(payload, fixture) {
			return fmt.Errorf("entry %d %q payload differs from fixture %q", index+1, file.Name, wanted.sameAs)
		}
	}
	return nil
}

func zipMetadata(file *zip.File, wanted zipEntryExpectation) error {
	if file.Name != wanted.name {
		return fmt.Errorf("name: expected %q, got %q", wanted.name, file.Name)
	}
	if file.Method != wanted.method {
		return fmt.Errorf("method: expected %d, got %d", wanted.method, file.Method)
	}
	creator := file.CreatorVersion >> 8
	if creator == 3 || creator == 19 {
		kind := (file.ExternalAttrs >> 16) & 0xf000
		if kind != 0 && kind != 0x8000 {
			return errors.New("type: entry is not a regular file")
		}
	}
	mode := file.Mode()
	if !mode.IsRegular() {
		return errors.New("type: entry is not a regular file")
	}
	if mode != wanted.mode {
		return fmt.Errorf("mode: expected %04o without special flags, got %v", wanted.mode, mode)
	}
	if !file.Modified.Equal(wanted.modified) {
		return fmt.Errorf("modified: expected %s, got %s", wanted.modified, file.Modified)
	}
	return nil
}

func zipPayload(file *zip.File, remaining int64) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	payload, readErr := io.ReadAll(io.LimitReader(reader, remaining+1))
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(payload)) > remaining {
		return nil, errors.New("aggregate actual uncompressed payload exceeds 1 MiB")
	}
	if crc32.ChecksumIEEE(payload) != file.CRC32 {
		return nil, errors.New("CRC checksum mismatch")
	}
	return payload, nil
}

func zipFixture(repositoryRoot, relative string, limit int64) ([]byte, error) {
	if !filepath.IsLocal(relative) {
		return nil, errors.New("fixture path must be local to the repository")
	}
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	target, err := filepath.EvalSymlinks(filepath.Join(root, relative))
	if err != nil {
		return nil, err
	}
	inside, err := filepath.Rel(root, target)
	if err != nil || !filepath.IsLocal(inside) {
		return nil, errors.New("fixture symlink target escapes the repository")
	}
	return readZIPRegular(target, limit)
}
