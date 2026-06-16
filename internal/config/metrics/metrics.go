package metrics

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
		host: envx.String("METRICS_HOST", ""),
		port: envx.String("METRICS_PORT", "9090"),
	}
}

func (c *Config) Address() string {
	return net.JoinHostPort(c.host, c.port)
}
