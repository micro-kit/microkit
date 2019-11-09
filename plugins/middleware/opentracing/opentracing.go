package opentracing

import (
	"context"
	"encoding/json"
	"io"
	slog "log"
	"strings"

	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

/* 链路追踪中间件 */

var (
	grpcTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)

// Opentracing 链路追踪中间件
type Opentracing struct {
	Options *Options
}

// NewOpentracing 创建链路追踪器
func NewOpentracing(opts ...Option) middleware.Middleware {
	opentracing := &Opentracing{
		Options: new(Options),
	}
	// 配置
	configure(opentracing, opts...)
	// 未设置日志对象退出
	if opentracing.Options.Logger == nil {
		slog.Fatalln("链路追踪中间件未设置日志对象")
	}

	return opentracing
}

// 配置设置项
func configure(zap *Opentracing, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(zap.Options)
	}
}

/* 服务端拦截器 */

// UnaryHandler 非流式中间件
func (trace *Opentracing) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if trace.Options.FilterOutFunc != nil && !trace.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}
	// 处理链路追踪数据
	newCtx, serverSpan := trace.newServerSpanFromInbound(ctx, trace.Options.Tracer, info.FullMethod)
	defer func() {
		if err != nil {
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(log.String("error", err.Error()))
			reqJs, _ := json.Marshal(req)
			serverSpan.LogFields(log.String("req", string(reqJs)))
			replyJs, _ := json.Marshal(resp)
			serverSpan.LogFields(log.String("resp", string(replyJs)))
		}
		serverSpan.Finish()
	}()

	// 执行下一步
	resp, err = handler(newCtx, req)
	return
}

// StreamHandler 流式中间件
func (trace *Opentracing) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if trace.Options.FilterOutFunc != nil && !trace.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}

	// 处理链路追踪数据
	_, serverSpan := trace.newServerSpanFromInbound(stream.Context(), trace.Options.Tracer, info.FullMethod)
	defer func() {
		if err != nil {
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(log.String("error", err.Error()))
		}
		serverSpan.Finish()
	}()

	err = handler(srv, stream)
	return
}

// MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	metadata.MD
}

// ForeachKey range all keys to call handler
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

// 服务端链路对象 https://segmentfault.com/a/1190000014546372
func (trace *Opentracing) newServerSpanFromInbound(ctx context.Context, tracer opentracing.Tracer, fullMethodName string) (context.Context, opentracing.Span) {
	//从context中取出metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	//从metadata中取出最终数据，并创建出span对象
	spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		trace.Options.Logger.Errorw("failed parsing trace information", "err", err)
	}
	serverSpan := tracer.StartSpan(
		fullMethodName,
		ext.RPCServerOption(spanContext),
		grpcTag,
		ext.SpanKindRPCServer,
	)
	// 这里在上下文记录了span对象 - 可通过opentracing.SpanFromContext(ctx)获取，可能为nil
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	return ctx, serverSpan
}

/* 以下客户端拦截器 */

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (trace *Opentracing) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if trace.Options.FilterOutFunc != nil && !trace.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}

	// 处理链路追踪数据
	newCtx, serverSpan := trace.newClientSpanFromContext(ctx, trace.Options.Tracer, method)
	defer func() {
		// 记录错误和请求响应参数
		if err != nil && err != io.EOF {
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(log.String("error", err.Error()))
			reqJs, _ := json.Marshal(req)
			serverSpan.LogFields(log.String("req", string(reqJs)))
			replyJs, _ := json.Marshal(reply)
			serverSpan.LogFields(log.String("resp", string(replyJs)))
		}
		serverSpan.Finish()
	}()

	// 执行下一步
	err = invoker(newCtx, method, req, reply, cc, opts...)
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (trace *Opentracing) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if trace.Options.FilterOutFunc != nil && !trace.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}

	// 处理链路追踪数据
	newCtx, clientSpan := trace.newClientSpanFromContext(ctx, trace.Options.Tracer, method)
	defer func() {
		if err != nil && err != io.EOF {
			ext.Error.Set(clientSpan, true)
			clientSpan.LogFields(log.String("error", err.Error()))
		}
		clientSpan.Finish()
	}()

	cs, err = streamer(newCtx, desc, cc, method, opts...)
	return
}

// 客户端链路对象 https://segmentfault.com/a/1190000014546372
func (trace *Opentracing) newClientSpanFromContext(ctx context.Context, tracer opentracing.Tracer, fullMethodName string) (context.Context, opentracing.Span) {
	// 从context中获取spanContext,如果上层没有开启追踪，则这里新建一个
	// 追踪，如果上层已经有了，测创建子span．
	var parentCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx = parent.Context()
	}
	cliSpan := tracer.StartSpan(
		fullMethodName,
		opentracing.ChildOf(parentCtx),
		grpcTag,
		ext.SpanKindRPCClient,
	)

	// 将之前放入context中的metadata数据取出，如果没有则新建一个metadata
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	mdWriter := MDReaderWriter{md}

	// 将追踪数据注入到metadata中
	err := tracer.Inject(cliSpan.Context(), opentracing.TextMap, mdWriter)
	if err != nil {
		trace.Options.Logger.Errorw("将追踪数据注入到metadata中错误", "err", err, "md", md, "fullMethodName", fullMethodName)
	}
	// 将metadata数据装入context中
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, cliSpan
}
