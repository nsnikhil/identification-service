package reporters

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
)

type Field struct {
	key   string
	value interface{}
}

func NewField(key string, value interface{}) Field {
	return Field{
		key:   key,
		value: key,
	}
}

type Logger interface {
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	InfoF(args ...interface{})
	DebugF(args ...interface{})
	WarnF(args ...interface{})
	ErrorF(args ...interface{})

	Flush() error

	Rotate() error
}

type zapLogger struct {
	lgr *zap.Logger
	lbr *lumberjack.Logger
}

func (zl *zapLogger) Info(msg string, fields ...Field) {
	zl.lgr.Info(msg, toZapFields(fields...)...)
}

func (zl *zapLogger) Debug(msg string, fields ...Field) {
	zl.lgr.Debug(msg, toZapFields(fields...)...)
}

func (zl *zapLogger) Warn(msg string, fields ...Field) {
	zl.lgr.Warn(msg, toZapFields(fields...)...)
}

func (zl *zapLogger) Error(msg string, fields ...Field) {
	zl.lgr.Error(msg, toZapFields(fields...)...)
}

func (zl *zapLogger) InfoF(args ...interface{}) {
	zl.lgr.Sugar().Info(args...)
}

func (zl *zapLogger) DebugF(args ...interface{}) {
	zl.lgr.Sugar().Debug(args...)
}

func (zl *zapLogger) WarnF(args ...interface{}) {
	zl.lgr.Sugar().Warn(args...)
}

func (zl *zapLogger) ErrorF(args ...interface{}) {
	zl.lgr.Sugar().Error(args...)
}

func toZapFields(fields ...Field) []zap.Field {
	sz := len(fields)

	res := make([]zap.Field, sz)

	for i := 0; i < sz; i++ {
		res[i] = zap.Any(fields[i].key, fields[i].value)
	}

	return res
}

func (zl *zapLogger) Flush() error {
	return zl.lgr.Sync()
}

func (zl *zapLogger) Rotate() error {
	if zl.lbr == nil {
		return nil
	}

	return zl.lbr.Rotate()
}

const production = "production"

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

func NewLogger(env, level string, writers ...io.Writer) Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig(env)),
		zapcore.NewMultiWriteSyncer(writeSyncers(writers...)...),
		zap.NewAtomicLevelAt(logLevel(level)),
	)

	return &zapLogger{
		lgr: zap.New(core),
		lbr: getLumberJack(writers...),
	}
}

//TODO: REMOVE CODE DUPLICATION
func getLumberJack(writers ...io.Writer) *lumberjack.Logger {
	for _, w := range writers {
		if _, ok := w.(*lumberjack.Logger); ok {
			return w.(*lumberjack.Logger)
		}
	}

	return nil
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
