package store

import (
	"strings"

	contract "onboard-kit/contracts/preparation/store"
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	preparationlocator "onboard-kit/internal/preparation/locator"

	"golang.org/x/sys/unix"
)

type selection struct {
	rootPath     string
	planPath     string
	rootIdentity fileIdentity
	plan         fileSnapshot
	planValue    model.Plan
}

func hasErrors(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func openSelection(rootPath, planPath string) (int, selection, contract.Code) {
	root, err := openDirectory(rootPath)
	if err != nil {
		return -1, selection{}, contract.CodePath
	}
	var stat unix.Stat_t
	if unix.Fstat(root, &stat) != nil {
		unix.Close(root)
		return -1, selection{}, contract.CodePath
	}
	plan, err := readPlanFile(planPath)
	if err != nil {
		unix.Close(root)
		return -1, selection{}, contract.CodeSource
	}
	value, problems := plandecoder.New().Decode(plan.content)
	if hasErrors(problems) {
		unix.Close(root)
		return -1, selection{}, contract.CodeSource
	}
	if value.ID == "" {
		unix.Close(root)
		return -1, selection{}, contract.CodePlan
	}
	return root, selection{rootPath: rootPath, planPath: planPath, rootIdentity: identityOf(stat), plan: plan, planValue: value}, ""
}

func locate(resultsPath, account string) (string, bool) {
	paths, problems := preparationlocator.New().Locate(resultsPath, account)
	return paths.Preparation, len(problems) == 0
}

func splitDestination(destination string) ([]string, string) {
	parts := strings.Split(destination, "/")
	return parts[:len(parts)-1], parts[len(parts)-1]
}

func readDestination(root int, destination string, planIdentity fileIdentity) (fileSnapshot, bool, contract.Code) {
	directories, name := splitDestination(destination)
	parent, err := walkDirectories(root, directories)
	if err == unix.ENOENT {
		return fileSnapshot{}, true, ""
	}
	if err != nil {
		return fileSnapshot{}, false, contract.CodePath
	}
	opened, err := openRegular(parent, name)
	unreadable := err == unix.EACCES && isRegularEntry(parent, name)
	unix.Close(parent)
	if err == unix.ENOENT {
		return fileSnapshot{}, false, ""
	}
	if unreadable {
		return fileSnapshot{}, false, contract.CodeSource
	}
	if err != nil || identityOf(opened.stat) == planIdentity {
		if err == nil {
			unix.Close(opened.descriptor)
		}
		return fileSnapshot{}, false, contract.CodePath
	}
	snapshot, err := readBounded(opened)
	if err != nil {
		return fileSnapshot{}, false, contract.CodeSource
	}
	return snapshot, false, ""
}
