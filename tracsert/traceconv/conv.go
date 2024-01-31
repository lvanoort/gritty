package traceconv

import (
	"github.com/lvanoort/gritty/internal/tprovider"
	"github.com/lvanoort/gritty/tracsert"
)

func ConvertGSToMatchableSet(spans ...*tprovider.GrittySpan) tracsert.SpanSet {
	var set tracsert.SpanSet
	for _, s := range spans {
		set = append(set, s)
	}

	return set
}
