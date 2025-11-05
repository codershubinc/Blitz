package utils

func LaunchApp(appName string) (string, error) {
	output, err := SpawnProcessBack(
		`gtk-launch`,
		[]string{appName},
	)
	if err != nil {
		return "", err
	}

	return output, nil
}
