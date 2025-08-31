package config

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
	Enable bool
	Users  map[string]string
}

func LoadConfig(path string) (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Host:       "0.0.0.0",
			Port:       1884,
			KeepAlive:  60,
			MaxClients: 1000,
		},
		Auth: AuthConfig{
			Enable: true,
			Users: map[string]string{
				"admin":    "password123",
				"ezlo_001": "password456",
				"ezlo_002": "password789"},
		},
	}, nil
}
