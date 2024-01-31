package matcher

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type MatchableSpan interface {
	GetName() string
	GetAttrs() []attribute.KeyValue
	GetStatusCode() codes.Code
	GetStatusDescription() string
}

type SpanMatcher interface {
	MatchSpan(span MatchableSpan) bool
}

type SpanMatcherFunc func(span MatchableSpan) bool

func (f SpanMatcherFunc) MatchSpan(span MatchableSpan) bool {
	return f(span)
}

type andMatcher struct {
	matchers []SpanMatcher
}

func (m andMatcher) MatchSpan(span MatchableSpan) bool {
	for _, m := range m.matchers {
		if !m.MatchSpan(span) {
			return false
		}
	}
	return true
}

func (m andMatcher) MarshalYAML() (any, error) {
	return nil, nil
}

func (m andMatcher) UnmarshalYAML(value any) error {
	return nil
}

func And(matchers ...SpanMatcher) SpanMatcher {
	return andMatcher{matchers: matchers}
}

type orMatcher struct {
	matchers []SpanMatcher
}

func (m orMatcher) MatchSpan(span MatchableSpan) bool {
	for _, m := range m.matchers {
		if m.MatchSpan(span) {
			return true
		}
	}
	return false
}

func Or(matchers ...SpanMatcher) SpanMatcher {
	return orMatcher{matchers: matchers}
}

func Not(matcher SpanMatcher) SpanMatcher {
	return SpanMatcherFunc(func(span MatchableSpan) bool {
		return !matcher.MatchSpan(span)
	})
}

// NameEquals tests if a span name equals something
// deprecated: this is temporary test code, the API sucks
func NameEquals(name string) SpanMatcher {
	return nameMatcher{name: name}
}

type nameMatcher struct {
	name string
}

func (m nameMatcher) MatchSpan(span MatchableSpan) bool {
	return span.GetName() == m.name
}
