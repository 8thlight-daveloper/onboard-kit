package create

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrExists = errors.New("the file already exists")

func WriteNew(path string, data []byte) error {
	if _, err := os.Lstat(path); err == nil {
		return ErrExists
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".obk-new-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
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
	if err := os.Chmod(temporary.Name(), 0o644); err != nil {
		return err
	}
	if err := os.Link(temporary.Name(), path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrExists
		}
		return err
	}
	return nil
}
