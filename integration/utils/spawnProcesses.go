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

func SpawnProcessBack(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)

	err := cmd.Start()
	if err != nil {
		return "", err
	}

	return "Command launched successfully. for =>  " + command, nil
}
