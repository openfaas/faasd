package tracer

import (
	_ "context"
	"io"
	_ "os"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	_ "github.com/opentracing/opentracing-go/log"
	_ "github.com/yurishkuro/opentracing-tutorial/go/lib/tracing"

	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
)

var lock = &sync.Mutex{}

type tracer struct {
	Tracer *opentracing.Tracer
	Closer *io.Closer
}

var tr *tracer

// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func IniTracer(service string) {
	/*
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6831",
		}
	*/
	if tr == nil {
		lock.Lock()
		defer lock.Unlock()
		if tr == nil {
			tr = &tracer{}
			cfg := &config.Configuration{
				ServiceName: service,
				Sampler: &config.SamplerConfig{
					Type:  "const",
					Param: 1,
				},
				Reporter: &config.ReporterConfig{
					LogSpans: false,
				},
			}
			tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
			if err != nil {
				panic("Can not init Jaeger")
			}
			tr.Tracer = &tracer
			tr.Closer = &closer
			opentracing.SetGlobalTracer(tracer)
		}
	}
}

func GetTracer() *tracer {
	if tr == nil {
		panic("Tracer is not Initialized")
	}
	return tr
}
