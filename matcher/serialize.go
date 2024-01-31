package matcher

var typeRegistry = make(map[string]SpanMatcher)

type marshallable struct {
	Type string `yaml:"type"`
	Args map[string]any
}
