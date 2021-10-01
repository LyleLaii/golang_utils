package zaplogger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
)

type ZapLogger struct {
	*zap.SugaredLogger
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

func ConfigZap(serverName string, runConf *RunningConfig) zap.Config {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	level := zapcore.Level(-1)
	switch runConf.Level.String() {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	atom := zap.NewAtomicLevelAt(level)

	dev := false
	if runConf.RunMode.String() == "dev" {
		dev = true
	}

	logFile := "./logs/alertmanager_notifier.log"

	ll := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, //MB
		MaxBackups: 5,
		MaxAge:     30, //days
		LocalTime:  true,
		Compress:   true,
	}
	zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: &ll,
		}, nil
	})

	config := zap.Config{
		Level:            atom,                                                // 日志级别
		Development:      dev,                                                 // 开发模式，堆栈跟踪
		Encoding:         "json",                                              // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                       // 编码器配置
		//InitialFields:    map[string]interface{}{"serviceName": serverName}, // 初始化字段
		//OutputPaths:      []string{"stdout"},         // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		//ErrorOutputPaths: []string{"stderr"},
		OutputPaths:      []string{"stdout", fmt.Sprintf("lumberjack:%s", logFile)},
		ErrorOutputPaths: []string{"stderr", fmt.Sprintf("lumberjack:%s", logFile)},
	}

	return config
}

func NewZapLogger(config zap.Config) *zap.Logger {
	logger, err := config.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

	return logger

}

func NewZapSugarLogger(config zap.Config) *ZapLogger {
	logger, err := config.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

	return &ZapLogger{logger.Sugar()}
}

func NewZapSugarLoggerGin(config zap.Config) *ZapLogger {
	logger, err := config.Build(zap.AddCallerSkip(9), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

	return &ZapLogger{logger.Sugar()}
}

func (l ZapLogger) Debug(module string, data interface{}) {
	l.With("module", module).Debug(data)
}

func (l ZapLogger) Info(module string, data interface{}) {
	l.With("module", module).Info(data)
}

func (l ZapLogger) Warn(module string, data interface{}) {
	l.With("module", module).Warn(data)
}

func (l ZapLogger) Error(module string, data interface{}) {
	l.With("module", module).Error(data)
}

func (l ZapLogger) Panic(module string, data interface{}) {
	l.With("module", module).Panic(data)
}

func (l ZapLogger) Fatal(module string, data interface{}) {
	l.With("module", module).Fatal(data)
}