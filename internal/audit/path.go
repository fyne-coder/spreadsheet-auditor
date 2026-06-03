package audit

import (
	"os"
	"os/user"
)

func userHomeDir() (string, error) {
	current, err := user.Current()
	if err == nil && current.HomeDir != "" {
		return current.HomeDir, nil
	}
	return os.UserHomeDir()
}
