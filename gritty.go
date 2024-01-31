package gritty

import (
	"context"
	"github.com/lvanoort/gritty/internal/tprovider"
	"github.com/lvanoort/gritty/matcher"
	"go.opentelemetry.io/otel"
	"testing"
)

type Gritty struct {
	t      *testing.T
	tracer *tprovider.GrittyTracerProvider
}

type grittyCtxKeyType string

var grittyCtxKey grittyCtxKeyType = "grittyCtx"

// parallelismCh lets us check if Gritty is being used in parallel, which is not
// supported due to dependency on manipulating the otel global tracer provider.
var parallelismCh = make(chan struct{}, 1)

func NewGritty(t *testing.T, ctx context.Context) (*Gritty, context.Context) {
	select {
	case parallelismCh <- struct{}{}:
		// good to go
	default:
		panic("Only one instance of Gritty can be in use at once")
	}

	originalTp := otel.GetTracerProvider()
	tp := tprovider.NewGrittyTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(originalTp)

		// pull from channel to allow next test to run
		<-parallelismCh
	})
	gritty := &Gritty{t, tp}

	return gritty, context.WithValue(ctx, grittyCtxKey, gritty)
}

func (g *Gritty) GetMatchingSpans(m matcher.SpanMatcher) []*tprovider.GrittySpan {
	spans := g.tracer.GetSpans()
	var matched []*tprovider.GrittySpan
	for _, s := range spans {
		if m.MatchSpan(s) {
			matched = append(matched, s)
		}
	}

	return matched
}
