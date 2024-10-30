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
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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

func CreateTar(prob string, byteCode []byte) (io.Reader, error) {
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

	if err := addCodeToTar(tw, byteCode); err != nil {
		return nil, err
	}

	path := path.Join("./internal/Content/Solution", prob)
	files := []string{"./judge.sh", path + "/input/input.txt", path + "/output/answer.txt"}
	for _, file := range files {
		if err := addFileToTar(tw, file); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func addCodeToTar(tw *tar.Writer, code []byte) error {
	header := &tar.Header{
		Name: "main.go",
		Size: int64(len(code)),
		Mode: 0600,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err := tw.Write(code)
	return err
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

func buildDockerImage(cli *client.Client, ctx context.Context, tarContext io.Reader, user string, prob string) (string, error) {
	imgName := "judge_system" + strings.ToLower(user) + strings.ToLower(prob) + ":latest"

	imageBuildResponse, err := cli.ImageBuild(ctx, tarContext, types.ImageBuildOptions{
		Context:    tarContext,
		Dockerfile: "Dockerfile",
		Tags:       []string{imgName},
		Remove:     true,
	})
	if err != nil {
		return "", err
	}
	defer imageBuildResponse.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(imageBuildResponse.Body)
	log.Println(buf.String())

	return imgName, nil
}

func runDockerContainer(cli *client.Client, ctx context.Context, repo string) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: repo,
	}, &container.HostConfig{
		Resources: container.Resources{
			// If U set resource lower than this, it will cause error (compilation error)
			Memory:   256 * 1024 * 1024, // 256MB
			NanoCPUs: 1000000000,        // 1 CPU
		},
	}, nil, nil, repo[:len(repo)-7])
	if err != nil {
		return "", err
	}
	log.Println(resp.ID)
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
	// Dangerous code use this carefully
	_, err = cli.ImageRemove(ctx, repo, image.RemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		return "", err
	}
	log.Println(buf.String())
	return buf.String(), nil
}
