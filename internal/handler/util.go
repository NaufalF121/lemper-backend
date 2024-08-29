package handler

import "os"

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
