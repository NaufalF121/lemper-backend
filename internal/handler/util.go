package handler

import (
	"os"
	"os/exec"
)

func saveStringToFile(filename, content string) error {
	// Create or open the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the string to the file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func execFile(filename string) (string, error) {
	cmd := exec.Command("go", "run", filename)

	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}
