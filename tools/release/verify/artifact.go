package main

import (
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func inspectArtifact(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("artifact must be an executable regular file: %s", path)
	}
	file, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read Mach-O artifact: %w", err)
	}
	defer file.Close()
	if file.Cpu != macho.CpuArm64 || file.Type != macho.TypeExec {
		return nil, fmt.Errorf("artifact must be a darwin/arm64 executable, got CPU %v, type %v", file.Cpu, file.Type)
	}
	libraries, err := file.ImportedLibraries()
	if err != nil {
		return nil, err
	}
	if err := checkLibraries(libraries); err != nil {
		return nil, err
	}
	return libraries, nil
}

func checkLibraries(libraries []string) error {
	for _, library := range libraries {
		switch library {
		case "/usr/lib/libSystem.B.dylib", "/usr/lib/libresolv.9.dylib":
		default:
			return fmt.Errorf("library outside the verified macOS baseline: %s", library)
		}
	}
	return nil
}

func fileHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("artifact must be a regular file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
