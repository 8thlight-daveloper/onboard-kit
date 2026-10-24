package editguide

import "os"

func inspect(file *os.File) (controls, error) { return controls{}, failNative() }

func supportedAttributes(values map[string]string) bool { return len(values) == 0 }

func establishAttributes(fd int, values map[string]string) error {
	if len(values) != 0 {
		return failNative()
	}
	return nil
}
