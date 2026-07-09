package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Token    string `json:"token"`
	UserName string `json:"user_name"`
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
func SaveToken(token string, userName string) error {
	path, err := getCfgPath()
	if err != nil {
		return err
	}

	cfg := Config{Token: token, UserName: userName}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadSession читает JWT из файла
func LoadSession() (*Config, error) {
	path, err := getCfgPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
