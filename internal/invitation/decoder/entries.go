package decoder

import (
	"archive/zip"
	"strings"

	"onboard-kit/internal/model"
)

type entry struct {
	file *zip.File
	role int
	path string
}

func identifyEntries(files []*zip.File) ([]entry, string, bool) {
	entries := make([]entry, 0, roleCount)
	seen := [roleCount]bool{}
	paths := [roleCount]string{"plan", "preparation", "record", "firstSteps"}
	account := ""
	for _, file := range files {
		role, currentAccount, valid := identifyName(file.Name)
		if !valid || seen[role] {
			return nil, "", false
		}
		if currentAccount != "" {
			if account != "" && account != currentAccount {
				return nil, "", false
			}
			account = currentAccount
		}
		seen[role] = true
		entries = append(entries, entry{file: file, role: role, path: paths[role]})
	}
	return entries, account, true
}

func identifyName(name string) (int, string, bool) {
	if len(name) > nameLimit {
		return 0, "", false
	}
	switch name {
	case "plan.obk.yaml":
		return 0, "", true
	case "first-steps.md":
		return 3, "", true
	}
	parts := strings.Split(name, "/")
	if len(parts) != 3 || parts[0] != "records" || !validAccount(parts[1]) {
		return 0, "", false
	}
	switch parts[2] {
	case "preparation.obk.yaml":
		return 1, parts[1], true
	case "record.obk.yaml":
		return 2, parts[1], true
	default:
		return 0, "", false
	}
}

func validAccount(account string) bool {
	if account == "" {
		return false
	}
	for index := range len(account) {
		letterOrDigit := account[index] >= 'A' && account[index] <= 'Z' ||
			account[index] >= 'a' && account[index] <= 'z' ||
			account[index] >= '0' && account[index] <= '9'
		punctuation := account[index] == '.' || account[index] == '_' || account[index] == '-'
		if !letterOrDigit && (index == 0 || !punctuation) {
			return false
		}
	}
	return true
}

func validateTypes(entries []entry) []model.Problem {
	for _, entry := range entries {
		if !regularFile(entry.file) {
			_, problems := failure("entry-type", "Recreate the invitation using regular files only.", entry.path)
			return problems
		}
	}
	return nil
}

func regularFile(file *zip.File) bool {
	creator := file.CreatorVersion >> 8
	if creator == 3 || creator == 19 {
		fileType := (file.ExternalAttrs >> 16) & 0xf000
		if fileType != 0 && fileType != 0x8000 {
			return false
		}
	}
	return file.Mode().IsRegular() && !strings.HasSuffix(file.Name, "/")
}

func validateEncodings(entries []entry) []model.Problem {
	const encryptionFlags = 1 | 1<<6 | 1<<13
	for _, entry := range entries {
		if entry.file.Flags&encryptionFlags != 0 {
			_, problems := failure("encrypted", "Recreate the invitation without ZIP encryption.", entry.path)
			return problems
		}
		if entry.file.Method != zip.Store && entry.file.Method != zip.Deflate {
			_, problems := failure("compression", "Recreate the invitation using Store or Deflate compression.", entry.path)
			return problems
		}
	}
	return nil
}

func withinDeclaredBudget(entries []entry) bool {
	remaining := uint64(payloadLimit)
	for _, entry := range entries {
		if entry.file.UncompressedSize64 > remaining {
			return false
		}
		remaining -= entry.file.UncompressedSize64
	}
	return true
}
