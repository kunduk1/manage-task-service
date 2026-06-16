// Package envx — общие хелперы чтения переменных окружения для реализаций конфигов.
package envx

import (
	"os"
	"strconv"
	"time"
)

// String возвращает значение переменной key, если она задана и непуста, иначе def.
func String(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// Int возвращает целочисленное значение переменной key, иначе def (в т.ч. при ошибке парсинга).
func Int(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// Duration возвращает значение переменной key как time.Duration, иначе def (в т.ч. при ошибке парсинга).
func Duration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// Bool возвращает булево значение переменной key, иначе def (в т.ч. при ошибке парсинга).
func Bool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
