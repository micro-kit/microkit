package common

import (
	"io"
	"time"

	"github.com/micro-kit/micro-common/config"
	opentracingGo "github.com/opentracing/opentracing-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

/* 生成一个链路追踪器 */

// NewJaegerTracer 生成一个默认链路追踪器
func NewJaegerTracer(serviceName string) (tracer opentracingGo.Tracer, closer io.Closer, err error) {
	cfg := jaegerCfg.Configuration{
		ServiceName: serviceName,
		RPCMetrics:  true,
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 3 * time.Second,
			LocalAgentHostPort:  config.GetJaegerAgentHostPort(),
		},
	}
	tracer, closer, err = cfg.NewTracer()
	if err != nil {
		return
	}
	opentracingGo.SetGlobalTracer(tracer)
	return
}
