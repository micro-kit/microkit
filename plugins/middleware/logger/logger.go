package logger

import (
	"context"

	"github.com/micro-kit/microkit/plugins/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/natefinch/lumberjack.v2"
)

/* 日志中间件 */

// ZapLogger zap 日志中间件
type ZapLogger struct {
	Options       *Options
	SugaredLogger *zap.SugaredLogger
}

// NewZapLogger 创建一个zap日志中间件
func NewZapLogger(filterOutFunc middleware.FilterFunc, opts ...Option) middleware.Middleware {
	zapLogger := &ZapLogger{
		Options: new(Options),
	}
	// 配置
	configure(zapLogger, opts...)
	// 未设置日志对象，则创建一个
	if zapLogger.Options.Logger == nil {
		// 创建zap日志对象
		syncWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:  zapLogger.Options.Filename,
			MaxSize:   int(zapLogger.Options.MaxSize),
			LocalTime: zapLogger.Options.LocalTime,
			Compress:  zapLogger.Options.Compress,
		})
		encoder := zap.NewProductionEncoderConfig()
		encoder.EncodeTime = zapcore.EpochMillisTimeEncoder // 时间格式
		core := zapcore.NewCore(zapcore.NewJSONEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(zapcore.Level(zapLogger.Options.Level)))
		logger := zap.New(core, zap.AddCaller())
		zapLogger.SugaredLogger = logger.Sugar()
	} else {
		zapLogger.SugaredLogger = zapLogger.Options.Logger
	}

	// 设置grpc日志对象
	grpclog.SetLoggerV2(&GrpcLog{
		SugaredLogger: zapLogger.SugaredLogger,
	})

	return zapLogger
}

// 配置设置项
func configure(zap *ZapLogger, ops ...Option) {
	// 默认值
	zap.Options.LocalTime = true
	zap.Options.Compress = true
	// 处理设置参数
	for _, o := range ops {
		o(zap.Options)
	}
	// 参数为空时默认值
	if zap.Options.Filename == "" {
		zap.Options.Filename = DefaultFilename
	}
	if zap.Options.MaxSize <= 0 {
		zap.Options.MaxSize = DefaultMaxSize
	}
	if zap.Options.Level < -1 || zap.Options.Level > 5 {
		zap.Options.Level = DefaultLevel
	}
}

// UnaryHandler 非流式中间件
func (zap *ZapLogger) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if zap.Options.FilterOutFunc != nil && !zap.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}
	// 记录日志
	defer func() {
		if err != nil {
			zap.SugaredLogger.Errorw("请求出现错误", "req", req, "resp", resp, "method", info.FullMethod, "err", err)
		} else if zap.Options.Level == zapcore.DebugLevel {
			zap.SugaredLogger.Debugw("请求日志", "req", req, "resp", resp, "method", info.FullMethod)
		}
	}()
	// 执行下一步
	resp, err = handler(ctx, req)
	return
}

// StreamHandler 流式中间件
func (zap *ZapLogger) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if zap.Options.FilterOutFunc != nil && !zap.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}
	defer func() {
		if err != nil {
			zap.SugaredLogger.Errorw("请求流函数", "method", info.FullMethod, "err", err)
		} else if zap.Options.Level == zapcore.DebugLevel {
			zap.SugaredLogger.Debugw("请求流函数", "method", info.FullMethod)
		}
	}()
	err = handler(srv, stream)
	return
}

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (zap *ZapLogger) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if zap.Options.FilterOutFunc != nil && !zap.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
	// 记录日志
	defer func() {
		if err != nil {
			zap.SugaredLogger.Errorw("请求出现错误", "req", req, "reply", reply, "method", method, "err", err)
		} else if zap.Options.Level == zapcore.DebugLevel {
			zap.SugaredLogger.Debugw("请求流函数", "method", method, "req", req, "reply", reply)
		}
	}()
	// 执行下一步
	err = invoker(ctx, method, req, reply, cc, opts...)
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (zap *ZapLogger) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if zap.Options.FilterOutFunc != nil && !zap.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}
	defer func() {
		if err != nil {
			zap.SugaredLogger.Errorw("请求流函数", "method", method, "err", err)
		} else if zap.Options.Level == zapcore.DebugLevel {
			zap.SugaredLogger.Debugw("请求流函数", "method", method, "err", err)
		}
	}()
	cs, err = streamer(ctx, desc, cc, method, opts...)
	return
}
