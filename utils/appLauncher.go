package utils

func LaunchApp(appName string) (string, error) {
	output, err := SpawnProcess(
		`gtk-launch`,
		[]string{appName},
	)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
