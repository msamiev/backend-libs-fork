package server

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"

	"github.com/labstack/gommon/log"
)

type EchoZapLogger struct {
	std     *zap.Logger
	core    zapcore.Core
	sugared *zap.SugaredLogger
}

// NewEchoZapLogger a logger that we can use for the echo server.
//
//	A few examples of how we can use this logger:
//	err := fmt.Errorf("test test")
//	ctx.Logger().Printf("1.Send message body: %s err: %s\n", body, err)
//	ctx.Logger().Info("2.Send message", zap.ByteString("msg", body), zap.Error(err))
//	ctx.Logger().Debug(zap.ByteString("msg", body), zap.Error(err))
func NewEchoZapLogger(std *zap.Logger) EchoZapLogger {
	std = std.WithOptions(zap.AddCallerSkip(1))
	return EchoZapLogger{
		std:     std,
		core:    std.Core(),
		sugared: std.Sugar(),
	}
}

func (l EchoZapLogger) Output() io.Writer {
	return &zapio.Writer{Log: l.std, Level: l.level()}
}

func (l EchoZapLogger) SetOutput(io.Writer) {}

// Prefix is not applicable to the zap.Logger and always returns an empty string
// because zap.Logger does not support any prefix in the messages.
func (l EchoZapLogger) Prefix() string {
	return ""
}

// SetPrefix is not applicable to the zap.Logger and does nothing
// because zap.Logger does not support any prefix in the messages.
func (l EchoZapLogger) SetPrefix(_ string) {}

func (l EchoZapLogger) Level() log.Lvl {
	switch l.level() { //nolint:exhaustive // default covers all cases
	case zapcore.DebugLevel:
		return log.DEBUG
	case zapcore.InfoLevel:
		return log.INFO
	case zapcore.WarnLevel:
		return log.WARN
	default:
		return log.ERROR
	}
}

func (l EchoZapLogger) level() zapcore.Level {
	switch {
	case l.core.Enabled(zapcore.DebugLevel):
		return zapcore.DebugLevel
	case l.core.Enabled(zapcore.InfoLevel):
		return zapcore.InfoLevel
	case l.core.Enabled(zapcore.WarnLevel):
		return zapcore.WarnLevel
	case l.core.Enabled(zapcore.ErrorLevel):
		return zapcore.ErrorLevel
	case l.core.Enabled(zapcore.DPanicLevel):
		return zapcore.DPanicLevel
	case l.core.Enabled(zapcore.PanicLevel):
		return zapcore.PanicLevel
	case l.core.Enabled(zapcore.FatalLevel):
		return zapcore.FatalLevel
	default:
		return zapcore.InvalidLevel
	}
}

// SetLevel is not applicable to the zap.Logger and does nothing
// because zap.Logger has already set the log level.
func (l EchoZapLogger) SetLevel(log.Lvl) {}

// SetHeader is not applicable to the zap.Logger and does nothing
// because zap.Logger does not support any header in the messages.
func (l EchoZapLogger) SetHeader(string) {}

func (l EchoZapLogger) Print(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Info(msg, fields...)
		return
	}
	l.sugared.Info(args...)
}

func (l EchoZapLogger) Printf(format string, args ...any) {
	l.sugared.Infof(format, args...)
}

func (l EchoZapLogger) Printj(j log.JSON) {
	l.sugared.Infow("json", j)
}

func (l EchoZapLogger) Debug(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Debug(msg, fields...)
		return
	}
	l.sugared.Debug(args...)
}

func (l EchoZapLogger) Debugf(format string, args ...any) {
	l.sugared.Debugf(format, args...)
}

func (l EchoZapLogger) Debugj(j log.JSON) {
	l.sugared.Debugw("json", j)
}

func (l EchoZapLogger) Info(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Info(msg, fields...)
		return
	}
	l.sugared.Info(args...)
}

func (l EchoZapLogger) Infof(format string, args ...any) {
	l.sugared.Infof(format, args...)
}

func (l EchoZapLogger) Infoj(j log.JSON) {
	l.sugared.Infow("json", j)
}

func (l EchoZapLogger) Warn(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Warn(msg, fields...)
		return
	}
	l.sugared.Warn(args...)
}

func (l EchoZapLogger) Warnf(format string, args ...any) {
	l.sugared.Warnf(format, args...)
}

func (l EchoZapLogger) Warnj(j log.JSON) {
	l.sugared.Warnw("json", j)
}

func (l EchoZapLogger) Error(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Error(msg, fields...)
		return
	}
	l.sugared.Error(args...)
}

func (l EchoZapLogger) Errorf(format string, args ...any) {
	l.sugared.Errorf(format, args...)
}

func (l EchoZapLogger) Errorj(j log.JSON) {
	l.sugared.Errorw("json", j)
}

func (l EchoZapLogger) Fatal(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Fatal(msg, fields...)
		return
	}
	l.sugared.Fatal(args...)
}

func (l EchoZapLogger) Fatalf(format string, args ...any) {
	l.sugared.Fatalf(format, args...)
}

func (l EchoZapLogger) Fatalj(j log.JSON) {
	l.sugared.Fatalw("json", j)
}

func (l EchoZapLogger) Panic(args ...any) {
	if msg, fields, ok := checkDefaultZapEntity(args...); ok {
		l.std.Panic(msg, fields...)
		return
	}
	l.sugared.Panic(args...)
}

func (l EchoZapLogger) Panicf(format string, args ...any) {
	l.sugared.Panicf(format, args...)
}

func (l EchoZapLogger) Panicj(j log.JSON) {
	l.sugared.Panicw("json", j)
}

func checkDefaultZapEntity(args ...any) (msg string, fields []zapcore.Field, ok bool) {
	if len(args) == 0 {
		return msg, fields, ok
	}

	fields = make([]zapcore.Field, 0)
	for i, arg := range args {
		if i == 0 {
			val, ok := arg.(string)
			if ok {
				msg = val
				continue
			}
		}

		if val, ok := arg.(zapcore.Field); ok {
			fields = append(fields, val)
			continue
		}

		return msg, fields, ok
	}

	return msg, fields, true
}
