package helloworld

import (
	"context"
	"github.com/lvanoort/gritty"
	"github.com/lvanoort/gritty/matcher"
	"github.com/lvanoort/gritty/tracsert"
	"github.com/lvanoort/gritty/tracsert/traceconv"
	"go.opentelemetry.io/otel/attribute"
	"net/http/httptest"
	"testing"
)

func TestRequest(t *testing.T) {
	ctx := context.Background()
	gr, ctx := gritty.NewGritty(t, ctx)

	req := httptest.NewRequest("GET", "/hello", nil).WithContext(ctx)
	reqID := "gritty-test-request-id"
	req.Header.Add("Internal-Request-ID", reqID)
	hw := NewHandler()
	hw.ServeHTTP(httptest.NewRecorder(), req)

	spans := gr.GetMatchingSpans(matcher.NameEquals("writeHello"))
	if len(spans) != 1 {
		t.Errorf("unexpected number of spans: %d", len(spans))
	}
	tracsert.ErrorTestOnAssertionFail(
		t,
		tracsert.SpanCount(1).Assert(traceconv.ConvertGSToMatchableSet(spans...)),
	)
	tracsert.ErrorTestOnAssertionFail(
		t,
		tracsert.AttributePresentAndEquals(
			attribute.String("internal.request.id", reqID)).Assert(traceconv.ConvertGSToMatchableSet(spans...)),
	)

}
