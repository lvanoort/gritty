package helloworld

import (
	"crypto/rand"
	"encoding/hex"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

type HelloWorld struct {
}

func NewHandler() *HelloWorld {
	return &HelloWorld{}
}

func (ls *HelloWorld) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	otelhttp.NewHandler(http.HandlerFunc(ls.serveHTTP), "greet").ServeHTTP(w, r)

}

var (
	writeTracer = otel.Tracer("write")
)

func (ls *HelloWorld) serveHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := writeTracer.Start(r.Context(), "writeHello")
	defer span.End()

	reqId := r.Header.Get("Internal-Request-ID")
	if reqId == "" {
		reqId = genNewReqID()
	}
	span.SetAttributes(attribute.String("internal.request.id", reqId))
	w.Write([]byte("Hello, world!"))
}

func genNewReqID() string {
	var data [16]byte
	_, _ = rand.Read(data[:])
	return hex.EncodeToString(data[:])
}
