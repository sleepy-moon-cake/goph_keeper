package add

import (
	"fmt"
	"goph_keeper/internal/shared/models"
	"io"
	"os"
)

var (
	textKey   string = name
	textValue string
	textPath  string = path
)

func handleText() (models.TextData, error) {
	var textData = models.TextData{Name: fileKey, Text: textValue}

	if textValue == "" && textPath == "" {
		return textData, fmt.Errorf("--value or --path should not be empty for --text flag")
	}

	if textPath != "" {
		fileInfo, err := os.Stat(textPath)
		if err != nil {
			return textData, fmt.Errorf("cant get stats: %w", err)
		}

		if int64(maxMemorySize) < fileInfo.Size() {
			return textData, fmt.Errorf("max file size 10MB")
		}

		file, err := os.OpenFile(textPath, os.O_RDONLY, 0)

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
