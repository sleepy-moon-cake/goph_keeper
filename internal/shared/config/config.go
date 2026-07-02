package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Token string `json:"token"`
}

func getCfgPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Путь вида: /Users/username/.gophkeeper.json
	return filepath.Join(home, ".gophkeeper.json"), nil
}

// SaveToken записывает JWT в файл
func SaveToken(token string) error {
	path, err := getCfgPath()
	if err != nil {
		return err
	}

	cfg := Config{Token: token}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadToken читает JWT из файла
func LoadToken() (string, error) {
	path, err := getCfgPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.Token, nil
}
