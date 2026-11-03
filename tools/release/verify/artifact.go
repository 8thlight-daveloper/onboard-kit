package main

import (
	"crypto/sha256"
	"debug/elf"
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

var linuxMachines = map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}

func inspectStaticArtifact(path string, machine elf.Machine) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return fmt.Errorf("artifact must be an executable regular file: %s", path)
	}
	file, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("read ELF artifact: %w", err)
	}
	defer file.Close()
	if file.Machine != machine || file.Type != elf.ET_EXEC {
		return fmt.Errorf("artifact must be a linux %v executable, got machine %v, type %v", machine, file.Machine, file.Type)
	}
	for _, program := range file.Progs {
		if program.Type == elf.PT_INTERP || program.Type == elf.PT_DYNAMIC {
			return fmt.Errorf("artifact must be statically linked, found program header %v", program.Type)
		}
	}
	libraries, err := file.ImportedLibraries()
	if err != nil {
		return err
	}
	if len(libraries) != 0 {
		return fmt.Errorf("artifact must link no shared libraries, found %v", libraries)
	}
	return nil
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
