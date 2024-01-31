package tracsert

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"slices"
	"strings"
	"testing"
)

type MatchableSpan interface {
	GetSpanID() trace.SpanID
	GetName() string
	GetAttrs() []attribute.KeyValue
	GetStatusCode() codes.Code
	GetStatusDescription() string
}

type SpanSet []MatchableSpan

type SpanSetAssertion interface {
	Assert(s SpanSet) AssertionResult
}

type SpanSetAssertionFunction func(s SpanSet) AssertionResult

func (s SpanSetAssertionFunction) Assert(set SpanSet) AssertionResult {
	return s(set)
}

type SpanAssertion interface {
	Assert(s SpanSet) AssertionResult
}

type SpanAssertionFunction func(s MatchableSpan) AssertionResult

func (s SpanAssertionFunction) Assert(set MatchableSpan) AssertionResult {
	return s(set)
}

type AssertionResult interface {
	Failed() bool
	Message() string
}

func ErrorTestOnAssertionFail(t *testing.T, result AssertionResult) {
	if result.Failed() {
		t.Error(result.Message())
	}
}

func FailTestOnAssertionFail(t *testing.T, result AssertionResult) {
	if result.Failed() {
		t.Fatal(result.Message())
	}
}

type strAssertion struct {
	failureMessage string
}

func (s strAssertion) Failed() bool {
	return s.failureMessage != ""
}

func (s strAssertion) Message() string {
	return s.failureMessage
}

type strAssertionFunction func(s SpanSet) strAssertion

func (s strAssertionFunction) Assert(set SpanSet) AssertionResult {
	return s(set)
}

func SpanCount(count int) SpanAssertion {
	return strAssertionFunction(func(s SpanSet) (res strAssertion) {
		if len(s) != count {
			res.failureMessage = fmt.Sprintf("Expected %d spans, got %d", count, len(s))
		}
		return
	})
}

func AttributeExists(key attribute.Key) SpanAssertion {
	return strAssertionFunction(func(s SpanSet) (res strAssertion) {
		for _, span := range s {
			for _, attr := range span.GetAttrs() {
				if attr.Key == key {
					return
				}
			}
			res.failureMessage = fmt.Sprintf("Expected attribute %s to exist on span %s, but it didn't", key, span.GetSpanID().String())
			return
		}
		return
	})
}

func AttributePresentAndEquals(desired attribute.KeyValue) SpanAssertion {
	return strAssertionFunction(func(s SpanSet) (res strAssertion) {
		for _, span := range s {
			for _, attr := range span.GetAttrs() {
				if attr.Key == desired.Key && valueEquals(attr.Value, desired.Value) {
					return
				}
			}
			res.failureMessage = fmt.Sprintf("Expected attribute %s to exist on span %s, but it didn't", string(desired.Key), span.GetSpanID().String())
			return
		}

		return
	})
}

func valueEquals(a, b attribute.Value) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch a.Type() {
	case attribute.INVALID:
		return false
	case attribute.BOOL:
		return a.AsBool() == b.AsBool()
	case attribute.INT64:
		return a.AsInt64() == b.AsInt64()
	case attribute.FLOAT64:
		return a.AsFloat64() == b.AsFloat64()
	case attribute.STRING:
		return a.AsString() == b.AsString()
	case attribute.BOOLSLICE:
		return slices.CompareFunc(a.AsBoolSlice(), b.AsBoolSlice(), func(b2 bool, b bool) int {
			if b2 == b {
				return 0
			} else if b {
				return 1
			} else {
				return -1
			}
		}) == 0
	case attribute.INT64SLICE:
		return slices.Compare(a.AsInt64Slice(), b.AsInt64Slice()) == 0
	case attribute.FLOAT64SLICE:
		return slices.Compare(a.AsFloat64Slice(), b.AsFloat64Slice()) == 0
	case attribute.STRINGSLICE:
		return slices.CompareFunc(a.AsStringSlice(), b.AsStringSlice(), strings.Compare) == 0
	}

	return true
}
