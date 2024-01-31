package tprovider

import (
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"time"
)

// idgen.go provides a very basic psuedo-random ID generator for use in testing. It seeds its PRNG
// in a predictable manner that depends on the current time, so it should not be used in production

type idGen interface {
	NewSpanID() trace.SpanID
	NewTraceID() trace.TraceID
}

type prngIDGen struct {
	rand *rand.Rand
}

var _ idGen = &prngIDGen{}

func (p *prngIDGen) NewSpanID() trace.SpanID {
	var sid trace.SpanID
	_, _ = p.rand.Read(sid[:])
	return sid
}

func (p *prngIDGen) NewTraceID() trace.TraceID {
	var tid trace.TraceID
	_, _ = p.rand.Read(tid[:])
	return tid
}

func newPrngIDGen() idGen {
	gen := &prngIDGen{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return gen
}
