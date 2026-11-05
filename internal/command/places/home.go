package places

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	ChosenByObkHome      = "OBK_HOME"
	ChosenByXDGStateHome = "XDG_STATE_HOME"
	ChosenByObkDataHome  = "OBK_DATA_HOME"
	ChosenByXDGDataHome  = "XDG_DATA_HOME"
	ChosenByDefault      = "the default"
)

type Home struct {
	Path   string
	Chosen string
}

func StateHome() (Home, error) {
	if fromEnvironment := os.Getenv("OBK_HOME"); fromEnvironment != "" {
		return Home{Path: fromEnvironment, Chosen: ChosenByObkHome}, nil
	}
	if runtime.GOOS != "darwin" {
		if state := os.Getenv("XDG_STATE_HOME"); state != "" {
			return Home{Path: filepath.Join(state, "obk"), Chosen: ChosenByXDGStateHome}, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return Home{}, err
		}
		return Home{Path: filepath.Join(home, ".local", "state", "obk"), Chosen: ChosenByDefault}, nil
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return Home{}, err
	}
	return Home{Path: filepath.Join(config, "obk"), Chosen: ChosenByDefault}, nil
}

const recordsFolderName = "onboarding"

func DataHome() (Home, error) {
	if fromEnvironment := os.Getenv("OBK_DATA_HOME"); fromEnvironment != "" {
		return Home{Path: fromEnvironment, Chosen: ChosenByObkDataHome}, nil
	}
	if fromEnvironment := os.Getenv("OBK_HOME"); fromEnvironment != "" {
		return Home{Path: fromEnvironment, Chosen: ChosenByObkHome}, nil
	}
	if runtime.GOOS != "darwin" {
		if data := os.Getenv("XDG_DATA_HOME"); data != "" {
			return Home{Path: filepath.Join(data, "obk"), Chosen: ChosenByXDGDataHome}, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return Home{}, err
		}
		return Home{Path: filepath.Join(home, ".local", "share", "obk"), Chosen: ChosenByDefault}, nil
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return Home{}, err
	}
	return Home{Path: filepath.Join(config, "obk"), Chosen: ChosenByDefault}, nil
}

func RecordsFolder(data Home) string {
	if data.Path == "" {
		return ""
	}
	return filepath.Join(data.Path, recordsFolderName)
}
