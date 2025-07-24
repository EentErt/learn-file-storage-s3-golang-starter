package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	args := []string{
		"-v", "error", "-print_format", "json", "-show_streams", filePath,
	}

	cmdOut := exec.Command("ffprobe", args...)
	buffer := bytes.NewBuffer([]byte{})
	cmdOut.Stdout = buffer

	var ffprobe struct {
		Streams struct {
			Width  int64 `json:"width"`
			Height int64 `json:"height"`
		} `json:"streams"`
	}

	if err := cmdOut.Run(); err != nil {
		return "", fmt.Errorf("unable to ffprobe file")
	}

	jsonFile, err := io.ReadAll(buffer)
	if err != nil {
		return "", fmt.Errorf("unable to read ffprobe output")
	}

	if err := json.Unmarshal(jsonFile, ffprobe); err != nil {
		return "", fmt.Errorf("unable to unmarshal json")
	}

	aspectRatio := ffprobe.Streams.Width / ffprobe.Streams.Height

	return fmt.Sprint(aspectRatio), nil
}
