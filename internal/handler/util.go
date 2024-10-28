package handler

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

func CreateTar(prob string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerfileContent := `
    FROM golang:1.23.1
    WORKDIR /app
    COPY main.go .
    COPY judge.sh .
    COPY input.txt .
    COPY answer.txt .
    RUN chmod +x judge.sh
    CMD ["./judge.sh"]
    `

	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfileContent)),
		Mode: 0600,
	}
	tw.WriteHeader(dockerfileHeader)
	tw.Write([]byte(dockerfileContent))

	path := path.Join("./internal/Content/Solution", prob)
	files := []string{"./internal/temp/main.go", "./judge.sh", path + "/input/input.txt", path + "/output/answer.txt"}
	for _, file := range files {
		if err := addFileToTar(tw, file); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func addFileToTar(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: filepath.Base(filename),
		Size: stat.Size(),
		Mode: 0600,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func buildDockerImage(cli *client.Client, ctx context.Context, tarContext io.Reader) (string, error) {
	imageBuildResponse, err := cli.ImageBuild(ctx, tarContext, types.ImageBuildOptions{
		Context:    tarContext,
		Dockerfile: "Dockerfile",
		Tags:       []string{"judge_system:latest"},
		Remove:     true,
	})
	if err != nil {
		return "", err
	}
	defer imageBuildResponse.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(imageBuildResponse.Body)
	log.Println(buf.String())

	return "judge_system:latest", nil
}

func runDockerContainer(cli *client.Client, ctx context.Context, image string) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:   128 * 1024 * 1024, // 256MB
			NanoCPUs: 500000000,         // 0.5 CPU
		},
	}, nil, nil, "judge")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); err != nil {
		panic(err)
	}
	return buf.String(), nil
}
