package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	Auth   AuthConfig
}

type ServerConfig struct {
	Host       string
	Port       int
	KeepAlive  int
	MaxClients int
}

type AuthConfig struct {
	Enabled              bool
	RequireAuth          bool
	AllowAnonymous       bool
	DefaultAnonymousUser string
}

func LoadConfig(path string) (*Config, error) {

	port := 1884
	if envPort := os.Getenv("GOBROMQ_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	host := "0.0.0.0"
	if envHost := os.Getenv("GOBROMQ_HOST"); envHost != "" {
		host = envHost
	}

	return &Config{
		Server: ServerConfig{
			Host:       host,
			Port:       port,
			KeepAlive:  60,
			MaxClients: 1000,
		},
		Auth: AuthConfig{
			Enabled:              true,
			RequireAuth:          true,
			AllowAnonymous:       false,
			DefaultAnonymousUser: "anonymous",
		},
	}, nil
}
