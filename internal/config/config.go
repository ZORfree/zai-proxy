package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	ProxyURL string
}

var Cfg *Config

func LoadConfig() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL == "" {
		proxyURL = "https://127.0.0.1:8989/api/random?protocol=socks5&max_latency=3000"
	}

	Cfg = &Config{
		Port:     port,
		ProxyURL: proxyURL,
	}
}
