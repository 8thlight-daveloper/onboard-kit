package places

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	ChosenByObkHome      = "OBK_HOME"
	ChosenByXDGStateHome = "XDG_STATE_HOME"
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
