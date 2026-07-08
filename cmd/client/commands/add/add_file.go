package add

import (
	"fmt"
	"goph_keeper/internal/shared/models"
	"io"
	"os"
)

func handleFile() (models.BinaryData, error) {
	fileData := models.BinaryData{Name: name}

	if path == "" {
		return fileData, fmt.Errorf("--path required for --file")
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return fileData, fmt.Errorf("cant get stats: %w", err)
	}

	if int64(maxMemorySize) < fileInfo.Size() {
		return fileData, fmt.Errorf("max file size 10MB")
	}

	fileData.FileName = fileInfo.Name()

	file, err := os.OpenFile(path, os.O_RDONLY, 0)

	if err != nil {
		return fileData, fmt.Errorf("cant open file: %w", err)
	}

	defer file.Close()

	byteData, err := io.ReadAll(file)

	if err != nil {
		return fileData, fmt.Errorf("cant read file: %w", err)
	}

	fileData.Data = byteData

	return fileData, nil
}
