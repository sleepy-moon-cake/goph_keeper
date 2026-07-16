package config

import (
	"flag"
	"os"
	"testing"
)

// Вспомогательная функция для сброса глобального состояния флагов перед каждым тестом
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

// Тест значений по умолчанию, когда флаги не переданы
func TestNewConfig_Defaults(t *testing.T) {
	resetFlags()

	// Имитируем запуск без аргументов (только имя бинарника)
	os.Args = []string{"cmd"}

	cfg := NewConfig()

	if cfg.ServerAddress != "localhost:8080" {
		t.Errorf("expected default ServerAddress 'localhost:8080', got '%s'", cfg.ServerAddress)
	}
	if cfg.GRPCServerAddress != ":3200" {
		t.Errorf("expected default GRPCServerAddress ':3200', got '%s'", cfg.GRPCServerAddress)
	}
	if cfg.DatabaseDSN != "" {
		t.Errorf("expected default DatabaseDSN to be empty, got '%s'", cfg.DatabaseDSN)
	}
	if cfg.SecretKey != "" {
		t.Errorf("expected default SecretKey to be empty, got '%s'", cfg.SecretKey)
	}
}

// Тест успешного парсинга всех переданных флагов
func TestNewConfig_WithFlags(t *testing.T) {
	resetFlags()

	// Имитируем передачу флагов через консоль
	os.Args = []string{
		"cmd",
		"-a", "127.0.0.1:9090",
		"-g", ":50051",
		"-d", "postgres://user:pass@localhost:5432/db",
		"-k", "my-super-secret-key",
	}

	cfg := NewConfig()

	if cfg.ServerAddress != "127.0.0.1:9090" {
		t.Errorf("expected ServerAddress '127.0.0.1:9090', got '%s'", cfg.ServerAddress)
	}
	if cfg.GRPCServerAddress != ":50051" {
		t.Errorf("expected GRPCServerAddress ':50051', got '%s'", cfg.GRPCServerAddress)
	}
	if cfg.DatabaseDSN != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("expected DatabaseDSN to match input, got '%s'", cfg.DatabaseDSN)
	}
	if cfg.SecretKey != "my-super-secret-key" {
		t.Errorf("expected SecretKey 'my-super-secret-key', got '%s'", cfg.SecretKey)
	}
}
