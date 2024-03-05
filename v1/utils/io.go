package utils

import (
	"fmt"
	"io"
	"os"
)

func ReadFileAsBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFileAsBytes: failed to open file: %w", err)
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ReadFileAsBytes: failed to read file: %w", err)
	}
	return content, nil
}
