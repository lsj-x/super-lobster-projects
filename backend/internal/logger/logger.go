package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func Init(level, output string) error {
	var err error
	var config zap.Config
	if level == "debug" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.Encoding = "json"
	config.OutputPaths = []string{output}
	Log, err = config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(Log)
	return nil
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Sync() {
	Log.Sync()
}
