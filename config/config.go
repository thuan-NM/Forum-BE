package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBDSN      string
	JWTSecret  string
	AllowAdmin bool
	ServerPort string
}

func LoadConfig() *Config {
	allowAdminStr := os.Getenv("ALLOW_ADMIN")
	allowAdmin, err := strconv.ParseBool(allowAdminStr)
	if err != nil {
		allowAdmin = false // Mặc định không cho phép admin nếu không thiết lập
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // Mặc định port 8080
	}

	return &Config{
		DBDSN:      os.Getenv("dataString"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AllowAdmin: allowAdmin,
		ServerPort: serverPort,
	}
}
