package config

import "flag"

type Config struct {
	ServerAddress     string
	GRPCServerAddress string
	DatabaseDSN       string
}

func NewConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.GRPCServerAddress, "g", ":3200", "gRPC server addresst")
	flag.StringVar(&config.DatabaseDSN, "d", "", "postgress url")

	flag.Parse()

	return &config
}
