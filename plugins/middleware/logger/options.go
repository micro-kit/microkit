package logger

import (
	"github.com/micro-kit/microkit/plugins/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/* 日志中间件参数 */

// Option 实例值设置
type Option func(*Options)

// 默认值
const (
	// DefaultLevel 默认日志级别
	DefaultLevel = zapcore.DebugLevel
	// DefaultFileName 默认日志输出路径
	DefaultFilename = "./access.log"
	// DefaultMaxSize 默认单日志文件大小
	DefaultMaxSize = 100
)

// 级别映射
var zapLevel = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

// Options 注册相关参数
type Options struct {
	Logger        *zap.SugaredLogger
	Filename      string
	MaxSize       int32
	LocalTime     bool
	Compress      bool
	Level         zapcore.Level
	FilterOutFunc middleware.FilterFunc
}

// Logger 设置日志对象
func Logger(logger *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// Filename 日志文件名
func Filename(filename string) Option {
	return func(o *Options) {
		o.Filename = filename
	}
}

// MaxSize 单日志文件大小上限
func MaxSize(maxSize int32) Option {
	return func(o *Options) {
		o.MaxSize = maxSize
	}
}

// LocalTime 是否使用本地事件
func LocalTime(localTime bool) Option {
	return func(o *Options) {
		o.LocalTime = localTime
	}
}

// Compress 是否压缩日志文件
func Compress(compress bool) Option {
	return func(o *Options) {
		o.Compress = compress
	}
}

// Level 日志级别
func Level(l string) Option {
	return func(o *Options) {
		if level, ok := zapLevel[l]; ok {
			o.Level = level
		} else {
			o.Level = DefaultLevel
		}
	}
}

// FilterOutFunc 设置中间件忽略函数列表
func FilterOutFunc(filterOutFunc middleware.FilterFunc) Option {
	return func(o *Options) {
		o.FilterOutFunc = filterOutFunc
	}
}
