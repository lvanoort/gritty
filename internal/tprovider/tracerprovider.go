package tprovider

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

func NewGrittyTracerProvider() *GrittyTracerProvider {
	return &GrittyTracerProvider{}
}

type GrittyTracerProvider struct {
	mu    sync.RWMutex
	spans []*GrittySpan
}

var _ trace.TracerProvider = &GrittyTracerProvider{}

func (p *GrittyTracerProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return &GrittyTracer{
		tp:   p,
		name: name,
		tc:   trace.NewTracerConfig(options...),
	}
}

func (p *GrittyTracerProvider) storeSpan(s *GrittySpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, s)
}

func (p *GrittyTracerProvider) GetSpans() []*GrittySpan {
	p.mu.RLock()
	defer p.mu.RUnlock()
	spanSlice := make([]*GrittySpan, 0, len(p.spans))
	for _, s := range p.spans {
		spanSlice = append(spanSlice, s)
	}

	return spanSlice
}

// GrittyTracer is a very basic OTel tracer for use in in-process trace-based integraiton testing. It is very basic and
// lacks a lot of features that a real tracer would have, but it's enough to get the job done for now.
// todo: add support for parent/child relationships
// todo: add support for sampling
type GrittyTracer struct {
	tp   *GrittyTracerProvider
	name string
	tc   trace.TracerConfig
}

var _ trace.Tracer = &GrittyTracer{}

func (g *GrittyTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	config := trace.NewSpanStartConfig(opts...)

	// good practice: always assume you _could_ be being passed a nil context, even if it seems like
	// something only a deranged lunatic would do
	if ctx == nil {
		ctx = context.Background()
	}

	var parentSpCtx trace.SpanContext
	if config.NewRoot() {
		// if we're establishing a new root, our parent context is a non-recording nil context
		ctx = trace.ContextWithSpanContext(ctx, parentSpCtx)
	} else {
		parentSpCtx = trace.SpanContextFromContext(ctx)
	}

	parentTraceID := parentSpCtx.TraceID()
	if !parentTraceID.IsValid() {
		parentTraceID = newPrngIDGen().NewTraceID()
	}

	spanID := newPrngIDGen().NewSpanID()

	scc := trace.SpanContextConfig{
		TraceID:    parentTraceID,
		SpanID:     spanID,
		TraceState: parentSpCtx.TraceState(),
		// todo: don't force sampling like this, but makes it easy to test for now
		TraceFlags: parentSpCtx.TraceFlags() | trace.FlagsSampled,
	}

	s := &GrittySpan{
		tp:        g.tp,
		startTime: config.Timestamp(),
		sc:        trace.NewSpanContext(scc),
		tracer:    g,
		name:      spanName,
		kind:      trace.ValidateSpanKind(config.SpanKind()),
		attrs:     append([]attribute.KeyValue{}, config.Attributes()...),
	}

	if s.startTime.IsZero() {
		s.startTime = time.Now()
	}

	g.tp.storeSpan(s)

	return trace.ContextWithSpan(ctx, s), s
}

// GrittySpan is a pretty basic span implementation for use in in-process trace-based integration testing. It is missing
// many features. Links for instance.
// todo: add Links
type GrittySpan struct {
	sc     trace.SpanContext
	tp     trace.TracerProvider
	tracer trace.Tracer

	startTime time.Time

	isEnded bool

	name  string
	attrs []attribute.KeyValue
	kind  trace.SpanKind

	recordedErrors []error
	recordedEvents []string
	recordedStatus *struct {
		code        codes.Code
		description string
	}

	lock sync.Mutex
}

var _ trace.Span = &GrittySpan{}

func (g *GrittySpan) End(options ...trace.SpanEndOption) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.isEnded = true
}

func (g *GrittySpan) AddEvent(name string, options ...trace.EventOption) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.recordedEvents = append(g.recordedEvents, name)
}

func (g *GrittySpan) IsRecording() bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	return true
}

func (g *GrittySpan) GetSpanID() trace.SpanID {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.sc.SpanID()
}

func (g *GrittySpan) RecordError(err error, options ...trace.EventOption) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.recordedErrors = append(g.recordedErrors, err)
}

func (g *GrittySpan) SpanContext() trace.SpanContext {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.sc
}

func (g *GrittySpan) SetStatus(code codes.Code, description string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.recordedStatus = &struct {
		code        codes.Code
		description string
	}{
		code:        code,
		description: description,
	}
}

func (g *GrittySpan) SetName(name string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.name = name
}

func (g *GrittySpan) SetAttributes(kv ...attribute.KeyValue) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.attrs = append(g.attrs, kv...)
}

func (g *GrittySpan) TracerProvider() trace.TracerProvider {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.tp
}

func (g *GrittySpan) GetName() string {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.name
}

func (g *GrittySpan) GetAttrs() []attribute.KeyValue {
	g.lock.Lock()
	defer g.lock.Unlock()

	attrCopy := make([]attribute.KeyValue, 0, len(g.attrs))
	for _, kv := range g.attrs {
		attrCopy = append(attrCopy, kv)
	}

	return attrCopy
}

func (g *GrittySpan) GetStatusCode() codes.Code {
	g.lock.Unlock()
	defer g.lock.Unlock()

	if g.recordedStatus == nil {
		return codes.Unset
	}
	return g.recordedStatus.code
}

func (g *GrittySpan) GetStatusDescription() string {
	g.lock.Unlock()
	defer g.lock.Unlock()

	if g.recordedStatus == nil {
		return ""
	}
	return g.recordedStatus.description
}
