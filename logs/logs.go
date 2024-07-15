package logs

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var LG *zap.Logger

type LogConfig struct {
	DebugFileName  string `yaml:"debugFileName" json:"debugFileName"`
	InfoFileName   string `yaml:"infoFileName" json:"infoFileName"`
	WarnFileName   string `yaml:"warnFileName" json:"warnFileName"`
	ErrorFileName  string `yaml:"errorFileName" json:"errorFileName"`
	DPanicFileName string `yaml:"dpanicFileName" json:"dpanicFileName"`
	PanicFileName  string `yaml:"panicFileName" json:"panicFileName"`
	FatalFileName  string `yaml:"fatalFileName" json:"fatalFileName"`
	MaxSize        int    `yaml:"maxsize" json:"maxsize"`
	MaxAge         int    `yaml:"max_age" json:"max_age"`
	MaxBackups     int    `yaml:"max_backups" json:"max_backups"`
}

// InitLogger 初始化Logger
func InitLogger(cfg *LogConfig) (err error) {
	writeSyncerDebug := getLogWriter(cfg.DebugFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerInfo := getLogWriter(cfg.InfoFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerWarn := getLogWriter(cfg.WarnFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerError := getLogWriter(cfg.ErrorFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerDPanic := getLogWriter(cfg.DPanicFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerPanic := getLogWriter(cfg.PanicFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerFatal := getLogWriter(cfg.FatalFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	//将log数据编码为json
	encoder := getEncoder()
	//文件输出
	debugCore := zapcore.NewCore(encoder, writeSyncerDebug, zapcore.DebugLevel)
	infoCore := zapcore.NewCore(encoder, writeSyncerInfo, zapcore.InfoLevel)
	warnCore := zapcore.NewCore(encoder, writeSyncerWarn, zapcore.WarnLevel)
	errorCore := zapcore.NewCore(encoder, writeSyncerError, zapcore.ErrorLevel)
	dpanicCore := zapcore.NewCore(encoder, writeSyncerDPanic, zapcore.DPanicLevel)
	panicCore := zapcore.NewCore(encoder, writeSyncerPanic, zapcore.PanicLevel)
	fatalCore := zapcore.NewCore(encoder, writeSyncerFatal, zapcore.FatalLevel)

	//标准输出
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	std := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
	core := zapcore.NewTee(debugCore, infoCore, warnCore, errorCore, dpanicCore, panicCore, fatalCore, std)
	LG = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(LG) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	//用于将时间戳格式化为 ISO 8601 标准的字符串表示形式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoderConfig.TimeKey = "time"
	//指定日志级别的编码方式
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//指定持续时间（Duration）的编码方式
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	//指定调用者（Caller）的编码方式
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	//return json格式的日志编码器
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}
