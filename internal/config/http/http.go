package http

import (
	"net"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	host string
	port string
}

func New() *Config {
	return &Config{
		host: envx.String("HTTP_HOST", ""),
		port: envx.String("HTTP_PORT", "8080"),
	}
}

func (c *Config) Address() string {
	return net.JoinHostPort(c.host, c.port)
}
