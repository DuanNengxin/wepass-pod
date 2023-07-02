package common

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func InitLogger(mode string) {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	var core zapcore.Core
	if mode == "debug" {
		consulEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			zapcore.NewCore(consulEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel),
		)
	} else {
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	}
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./wepass-base.log",
		MaxSize:    50,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
