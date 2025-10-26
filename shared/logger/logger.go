package logger

import "go.uber.org/zap"

var globalLogger *zap.Logger

func Init(env string) error{
	var logger *zap.Logger
	var err error
	
	if env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

func Get() *zap.Logger {
	if globalLogger == nil {
		globalLogger = zap.NewNop()
	}
	return globalLogger
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}
