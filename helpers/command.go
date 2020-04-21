package helpers

import (
	"os/exec"

	"github.com/google/logger"
)

// GetCommandOutput runs a command in a subshell and returns the output as a string
func GetCommandOutput(logger *logger.Logger, command string, args []string) (output string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("cmd.Run() failed with %s\n", err)
		return
	}
	sha := string(out)

	return sha
}
