package add

import (
	"fmt"
	"goph_keeper/internal/shared/models"
	"io"
	"os"
)

func handleText() (models.TextData, error) {
	var textData = models.TextData{Name: name, Text: value}

	if value == "" && path == "" {
		return textData, fmt.Errorf("--value or --path should not be empty for --text flag")
	}

	if path != "" {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return textData, fmt.Errorf("cant get stats: %w", err)
		}

		if int64(maxMemorySize) < fileInfo.Size() {
			return textData, fmt.Errorf("max file size 10MB")
		}

		file, err := os.OpenFile(path, os.O_RDONLY, 0)

		if err != nil {
			return textData, fmt.Errorf("cant open file: %w", err)
		}

		defer file.Close()

		byteData, err := io.ReadAll(file)

		if err != nil {
			return textData, fmt.Errorf("cant read file: %w", err)
		}

		textData.Text = string(byteData)
	}

	fmt.Println(textData)

	return textData, nil
}
