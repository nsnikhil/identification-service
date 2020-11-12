package reporters

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

const production = "production"

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

func NewLogger(env, level string, writers ...io.Writer) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig(env)),
		zapcore.NewMultiWriteSyncer(writeSyncers(writers...)...),
		zap.NewAtomicLevelAt(logLevel(level)),
	)

	return zap.New(core)
}

func writeSyncers(writers ...io.Writer) []zapcore.WriteSyncer {
	var res []zapcore.WriteSyncer
	for _, w := range writers {
		res = append(res, zapcore.AddSync(w))
	}

	return res
}

func logLevel(level string) zapcore.Level {
	l, ok := levelMap[level]
	if !ok {
		return zapcore.InfoLevel
	}

	return l
}

//TODO: IS THE PREBUILT ENCODER CONFIG GOOD ENOUGH? OR DO YOU NEED TO A CUSTOM ONE?
func encoderConfig(env string) zapcore.EncoderConfig {
	if env == production {
		return zap.NewProductionEncoderConfig()
	}

	return zap.NewDevelopmentEncoderConfig()
}
