package store

import (
	"os"
	"sort"
	"strings"

	contract "onboard-kit/contracts/preparation/store"

	"golang.org/x/sys/unix"
)

const (
	accountToken   = "{account}"
	probeAccount   = "probe"
	defaultResults = "records/{account}/record.obk.yaml"
)

func (store) List(request contract.ListRequest) contract.ListResult {
	root, selected, code := openSelection(request.Root, request.Plan)
	if code != "" {
		return contract.ListResult{Code: code}
	}
	defer unix.Close(root)
	resultsPath := selected.planValue.ResultsPath
	if resultsPath == "" {
		resultsPath = defaultResults
	}
	if _, ok := locate(resultsPath, probeAccount); !ok {
		return contract.ListResult{Code: contract.CodePath}
	}
	components := strings.Split(resultsPath[:strings.LastIndex(resultsPath, "/")], "/")
	tokenIndex := 0
	for !strings.Contains(components[tokenIndex], accountToken) {
		tokenIndex++
	}
	holder, err := walkDirectories(root, components[:tokenIndex])
	if err == unix.ENOENT {
		return contract.ListResult{Code: contract.CodeListed}
	}
	if err != nil {
		return contract.ListResult{Code: contract.CodePath}
	}
	defer unix.Close(holder)
	names, err := directoryNames(holder)
	if err != nil {
		return contract.ListResult{Code: contract.CodePath}
	}
	sort.Strings(names)
	var people []contract.Person
	for _, name := range names {
		account, matches := accountInName(components[tokenIndex], name)
		if !matches {
			continue
		}
		destination, ok := locate(selected.planValue.ResultsPath, account)
		if !ok {
			continue
		}
		if problem, listed := preparationPresence(holder, destination, tokenIndex); listed {
			people = append(people, contract.Person{Account: account, Destination: destination, Problem: problem})
		}
	}
	return contract.ListResult{Code: contract.CodeListed, People: people}
}

func directoryNames(descriptor int) ([]string, error) {
	duplicate, err := unix.Openat(descriptor, ".", directoryFlags, 0)
	if err != nil {
		return nil, err
	}
	directory := os.NewFile(uintptr(duplicate), "listed-directory")
	names, readErr := directory.Readdirnames(-1)
	closeErr := directory.Close()
	if readErr != nil {
		return nil, readErr
	}
	return names, closeErr
}

func accountInName(pattern, name string) (string, bool) {
	count := strings.Count(pattern, accountToken)
	extra := len(name) - (len(pattern) - count*len(accountToken))
	if extra <= 0 || extra%count != 0 {
		return "", false
	}
	start := strings.Index(pattern, accountToken)
	account := name[start : start+extra/count]
	return account, strings.ReplaceAll(pattern, accountToken, account) == name
}

func preparationPresence(holder int, destination string, tokenIndex int) (contract.Code, bool) {
	directories, name := splitDestination(destination)
	parent, err := walkDirectories(holder, directories[tokenIndex:])
	if err != nil {
		return "", false
	}
	defer unix.Close(parent)
	opened, err := openRegular(parent, name)
	if err == unix.EACCES {
		return contract.CodeSource, isRegularEntry(parent, name)
	}
	if err != nil {
		return "", false
	}
	return "", unix.Close(opened.descriptor) == nil
}

func isRegularEntry(parent int, name string) bool {
	var stat unix.Stat_t
	return unix.Fstatat(parent, name, &stat, unix.AT_SYMLINK_NOFOLLOW) == nil && uint32(stat.Mode)&unix.S_IFMT == unix.S_IFREG
}
