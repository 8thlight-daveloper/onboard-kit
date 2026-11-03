package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"go/build"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	cacheFolder    = ".direnv/mutation-cache"
	gomuVersion    = "v0.3.0"
	cacheSeparator = "\x00"
)

func sourceFiles(root string, component string) ([]string, error) {
	built, _, err := splitSourceFiles(root, component)
	return built, err
}

func splitSourceFiles(root string, component string) ([]string, []string, error) {
	folder := filepath.Join(root, "internal", component)
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, nil, err
	}
	var built, skipped []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if matches, err := build.Default.MatchFile(folder, name); err == nil && matches {
			built = append(built, name)
		} else {
			skipped = append(skipped, name)
		}
	}
	return built, skipped, nil
}

func suiteDigest(root string, component string) (string, error) {
	base := filepath.Join(root, "contracts", component)
	var paths []string
	err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			paths = append(paths, path)
		}
		return err
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		hash.Write([]byte(strings.TrimPrefix(path, base) + cacheSeparator))
		hash.Write(data)
		hash.Write([]byte(cacheSeparator))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func cacheKey(source []byte, suite string, short bool) string {
	hash := sha256.New()
	hash.Write(source)
	hash.Write([]byte(cacheSeparator + suite + cacheSeparator + gomuVersion))
	if short {
		hash.Write([]byte(cacheSeparator + "short"))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func cachePath(root string, component string, name string, key string) string {
	folder := strings.ReplaceAll(component, "/", "_")
	return filepath.Join(root, cacheFolder, folder, name+"."+key+".json")
}

func readCached(path string) ([]result, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var results []result
	if json.Unmarshal(data, &results) != nil {
		return nil, false
	}
	return results, true
}

func writeCached(path string, results []result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func forgetStale(root string, component string, name string, keep string) {
	pattern := cachePath(root, component, name, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil && !errors.Is(err, filepath.ErrBadPattern) {
		return
	}
	for _, match := range matches {
		if match != keep {
			os.Remove(match)
		}
	}
}
