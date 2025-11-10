package utils

import (
	"os/exec"
)

func SpawnProcess(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}
