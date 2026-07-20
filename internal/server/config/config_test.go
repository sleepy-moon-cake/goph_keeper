package config

import (
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestNewConfig_Defaults(t *testing.T) {
	resetFlags()

	os.Args = []string{"cmd"}

	cfg, err := NewConfig()

	if err == nil {
		t.Fatal("expected error due to empty secret key by default, got nil")
	}
	if cfg != nil {
		t.Errorf("expected config to be nil on error, got %v", cfg)
	}
}

func TestNewConfig_WithFlags(t *testing.T) {
	resetFlags()

	os.Args = []string{
		"cmd",
		"-a", "127.0.0.1:9090",
		"-g", ":50051",
		"-d", "postgres://user:pass@localhost:5432/db",
		"-k", "my-super-secret-key",
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestNewConfig_MissingSecretKey(t *testing.T) {
	resetFlags()

	os.Args = []string{
		"cmd",
		"-a", "localhost:8080",
		"-g", ":3200",
		"-d", "postgres://localhost:5432",
	}

	cfg, err := NewConfig()
	if err == nil {
		t.Fatal("expected error due to missing secret key (-k), got nil")
	}
	if cfg != nil {
		t.Errorf("expected config to be nil on error, got %v", cfg)
	}
}
