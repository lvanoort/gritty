package gritty

import (
	"context"
	"testing"
)

func TestNewGritty(t *testing.T) {
	g, ctx := NewGritty(t, context.Background())
	if ctx == nil {
		t.Error("context is nil")
	}
	if g == nil {
		t.Error("returned gritty is nil")
	}

	gritty := ctx.Value(grittyCtxKey)
	if gritty == nil {
		t.Error("gritty is nil in context")
	}
}

func TestNewGrittyParallel(t *testing.T) {
	NewGritty(t, context.Background())
	success := false
	fn := func() {
		NewGritty(t, context.Background())
	}
	testFn := func() {
		defer func() {
			if r := recover(); r != nil {
				success = true
			}
		}()
		fn()
	}

	testFn()
	if !success {
		t.Error("expected panic")
	}
}
