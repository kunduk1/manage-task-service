package mysql

import (
	"fmt"
	"net"
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	host            string
	port            string
	user            string
	password        string
	database        string
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

func New() *Config {
	return &Config{
		host:            envx.String("MYSQL_HOST", "localhost"),
		port:            envx.String("MYSQL_PORT", "3306"),
		user:            envx.String("MYSQL_USER", "root"),
		password:        envx.String("MYSQL_PASSWORD", ""),
		database:        envx.String("MYSQL_DATABASE", "manage_task"),
		maxOpenConns:    envx.Int("MYSQL_MAX_OPEN_CONNS", 25),
		maxIdleConns:    envx.Int("MYSQL_MAX_IDLE_CONNS", 25),
		connMaxLifetime: envx.Duration("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC",
		c.user, c.password, net.JoinHostPort(c.host, c.port), c.database,
	)
}

func (c *Config) MaxOpenConns() int { return c.maxOpenConns }

func (c *Config) MaxIdleConns() int { return c.maxIdleConns }

func (c *Config) ConnMaxLifetime() time.Duration { return c.connMaxLifetime }
