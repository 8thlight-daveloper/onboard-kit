package registry

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	FileName       = "plans.txt"
	FirstLine      = "onboard-kit plans 1"
	MaximumBytes   = 256 << 10
	MaximumEntries = 100
	MaximumPath    = 4096
	timeLayout     = time.RFC3339
	fieldSeparator = "\t"
)

type Entry struct {
	Path   string
	Opened time.Time
}

type Damage struct {
	File   string
	Reason string
}

func (damage *Damage) Error() string {
	return fmt.Sprintf("the list of plans at %s cannot be read: %s", damage.File, damage.Reason)
}

var ErrUnfitPath = errors.New("the plan path is not a full path of plain text")

func Path(home string) string { return filepath.Join(home, FileName) }

func Load(home string) ([]Entry, error) {
	file, err := os.Open(Path(home))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaximumBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaximumBytes {
		return nil, &Damage{File: Path(home), Reason: fmt.Sprintf("it is larger than %d KiB", MaximumBytes>>10)}
	}
	entries, reason := Parse(data)
	if reason != "" {
		return nil, &Damage{File: Path(home), Reason: reason}
	}
	return entries, nil
}

func Parse(data []byte) ([]Entry, string) {
	if !utf8.Valid(data) {
		return nil, "it is not UTF-8 text"
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if rows[0] != FirstLine {
		return nil, "its first line is not " + FirstLine
	}
	var entries []Entry
	seen := map[string]bool{}
	for index, row := range rows[1:] {
		if strings.TrimSpace(row) == "" {
			continue
		}
		number := index + 2
		stamp, path, found := strings.Cut(row, fieldSeparator)
		opened, err := time.Parse(timeLayout, stamp)
		if !found || err != nil || !fitPath(path) {
			return nil, fmt.Sprintf("line %d is not a time, a tab, and a full path", number)
		}
		if seen[path] {
			return nil, fmt.Sprintf("line %d names a plan a second time", number)
		}
		seen[path] = true
		entries = append(entries, Entry{Path: path, Opened: opened.UTC()})
		if len(entries) > MaximumEntries {
			return nil, fmt.Sprintf("it holds more than %d plans", MaximumEntries)
		}
	}
	return newestFirst(entries), ""
}

func Format(entries []Entry) []byte {
	var text strings.Builder
	text.WriteString(FirstLine + "\n")
	for _, entry := range entries {
		text.WriteString(entry.Opened.UTC().Format(timeLayout) + fieldSeparator + entry.Path + "\n")
	}
	return []byte(text.String())
}

func Record(home string, path string, opened time.Time) error {
	if !fitPath(path) {
		return ErrUnfitPath
	}
	entries, err := Load(home)
	if err != nil {
		return err
	}
	kept := []Entry{{Path: path, Opened: opened.UTC().Truncate(time.Second)}}
	for _, entry := range entries {
		if entry.Path != path && len(kept) < MaximumEntries {
			kept = append(kept, entry)
		}
	}
	return replace(home, Format(newestFirst(kept)))
}

func Forget(home string, path string) error {
	entries, err := Load(home)
	if err != nil {
		return err
	}
	kept := slices.DeleteFunc(slices.Clone(entries), func(entry Entry) bool { return entry.Path == path })
	if len(kept) == len(entries) {
		return nil
	}
	return replace(home, Format(kept))
}

func newestFirst(entries []Entry) []Entry {
	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(first Entry, second Entry) int { return second.Opened.Compare(first.Opened) })
	return sorted
}

func fitPath(path string) bool {
	if path == "" || len(path) > MaximumPath || !utf8.ValidString(path) || !filepath.IsAbs(path) {
		return false
	}
	return !strings.ContainsFunc(path, unicode.IsControl)
}

func replace(home string, data []byte) error {
	if err := os.MkdirAll(home, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(home, ".plans-*.tmp")
	if err != nil {
		return err
	}
	written := false
	defer func() {
		if !written {
			os.Remove(temporary.Name())
		}
	}()
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary.Name(), Path(home)); err != nil {
		return err
	}
	written = true
	return nil
}
