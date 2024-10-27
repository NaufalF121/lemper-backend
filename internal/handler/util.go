package handler

import (
	"bufio"
	"io"
	"os"
	"os/exec"
)

func saveStringToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func execFile(filename string, input []string) (string, error) {
	var echo string
	for _, line := range input {
		echo += line + "\n"
	}
	r, w := io.Pipe()
	echoCmd := exec.Command("echo", echo)
	echoCmd.Stdout = w
	runCmd := exec.Command("go", "run", filename)
	runCmd.Stdin = r
	if err := echoCmd.Start(); err != nil {
		return "", err
	}
	out, err := runCmd.Output()
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parseTxtFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
