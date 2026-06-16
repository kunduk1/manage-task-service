package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var golbalLogger *zap.Logger

func NewGlobalLogger(core zapcore.Core, opts ...zap.Option) {
	golbalLogger = zap.New(core, opts...)
}

func Info(msg string, fields ...zap.Field) {
	golbalLogger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	golbalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	golbalLogger.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	golbalLogger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	golbalLogger.Warn(msg, fields...)
}

func WithOptions(opts ...zap.Option) *zap.Logger {
	return golbalLogger.WithOptions(opts...)
}
