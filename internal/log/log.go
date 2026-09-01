package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	// Format methods
	Printf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	// Line methods (like logrus Infoln)
	Println(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})

	// Field methods - using standard Go types, not zap.Field
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Sugar() *zap.SugaredLogger
}

type Fields map[string]interface{}

type Config struct {
	LogLevel string
}

type ZapAdapter struct {
	logger *zap.SugaredLogger
}

func New(config *Config) *ZapAdapter {
	// Get the log level from the config
	var level zapcore.Level
	switch config.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create a new production config
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Build the logger
	zapLogger, err := zapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		zapLogger = zap.NewNop()
	}
	return &ZapAdapter{
		logger: zapLogger.Sugar(),
	}
}

func (l *ZapAdapter) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *ZapAdapter) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *ZapAdapter) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *ZapAdapter) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *ZapAdapter) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *ZapAdapter) Printf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *ZapAdapter) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *ZapAdapter) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *ZapAdapter) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *ZapAdapter) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *ZapAdapter) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *ZapAdapter) Println(args ...interface{}) {
	l.logger.Debug(fmt.Sprintln(args...))
}

func (l *ZapAdapter) Debugln(args ...interface{}) {
	l.logger.Debug(fmt.Sprintln(args...))
}
func (l *ZapAdapter) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprintln(args...))
}
func (l *ZapAdapter) Warnln(args ...interface{}) {
	l.logger.Warn(fmt.Sprintln(args...))
}
func (l *ZapAdapter) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprintln(args...))
}
func (l *ZapAdapter) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprintln(args...))
}

// Field methods
func (z *ZapAdapter) WithField(key string, value interface{}) Logger {
	return &ZapAdapter{
		logger: z.logger.With(key, value),
	}
}

func (z *ZapAdapter) WithFields(fields Fields) Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &ZapAdapter{
		logger: z.logger.With(args...),
	}
}

func (z *ZapAdapter) WithError(err error) Logger {
	return &ZapAdapter{
		logger: z.logger.With("error", err),
	}
}

func (z *ZapAdapter) Sugar() *zap.SugaredLogger {
	return z.logger
}
