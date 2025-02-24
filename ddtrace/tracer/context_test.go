// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zleague/dd-trace-go/ddtrace/internal"
)

func TestContextWithSpan(t *testing.T) {
	want := &span{SpanID: 123}
	ctx := ContextWithSpan(context.Background(), want)
	got, ok := ctx.Value(activeSpanKey).(*span)
	assert := assert.New(t)
	assert.True(ok)
	assert.Equal(got, want)
}

func TestSpanFromContext(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		assert := assert.New(t)
		want := &span{SpanID: 123}
		ctx := ContextWithSpan(context.Background(), want)
		got, ok := SpanFromContext(ctx)
		assert.True(ok)
		assert.Equal(got, want)
	})
	t.Run("no-op", func(t *testing.T) {
		assert := assert.New(t)
		span, ok := SpanFromContext(context.Background())
		assert.False(ok)
		_, ok = span.(*internal.NoopSpan)
		assert.True(ok)
		span, ok = SpanFromContext(nil)
		assert.False(ok)
		_, ok = span.(*internal.NoopSpan)
		assert.True(ok)
	})
}

func TestStartSpanFromContext(t *testing.T) {
	_, _, _, stop := startTestTracer(t)
	defer stop()

	parent := &span{context: &spanContext{spanID: 123, traceID: 456}}
	parent2 := &span{context: &spanContext{spanID: 789, traceID: 456}}
	pctx := ContextWithSpan(context.Background(), parent)
	child, ctx := StartSpanFromContext(
		pctx,
		"http.request",
		ServiceName("gin"),
		ResourceName("/"),
		ChildOf(parent2.Context()), // we do this to assert that the span in pctx takes priority.
	)
	assert := assert.New(t)

	got, ok := child.(*span)
	assert.True(ok)
	gotctx, ok := SpanFromContext(ctx)
	assert.True(ok)
	assert.Equal(gotctx, got)
	_, ok = gotctx.(*internal.NoopSpan)
	assert.False(ok)

	assert.Equal(uint64(456), got.TraceID)
	assert.Equal(uint64(123), got.ParentID)
	assert.Equal("http.request", got.Name)
	assert.Equal("gin", got.Service)
	assert.Equal("/", got.Resource)
}

func TestStartSpanFromContextRace(t *testing.T) {
	_, _, _, stop := startTestTracer(t)
	defer stop()

	// Start 100 goroutines that create child spans with StartSpanFromContext in parallel,
	// with a shared options slice. The child spans should get parented to the correct spans
	const contextKey = "key"
	const numContexts = 100
	options := make([]StartSpanOption, 0, 3)
	outputValues := make(chan uint64, numContexts)
	var expectedTraceIDs []uint64
	for i := 0; i < numContexts; i++ {
		parent, childCtx := StartSpanFromContext(context.Background(), "parent")
		expectedTraceIDs = append(expectedTraceIDs, parent.Context().TraceID())
		go func() {
			span, _ := StartSpanFromContext(childCtx, "testoperation", options...)
			defer span.Finish()
			outputValues <- span.Context().TraceID()
		}()
		parent.Finish()
	}

	// collect the outputs
	var outputs []uint64
	for i := 0; i < numContexts; i++ {
		outputs = append(outputs, <-outputValues)
	}
	assert.Len(t, outputs, numContexts)
	assert.ElementsMatch(t, outputs, expectedTraceIDs)
}

func TestStartSpanFromNilContext(t *testing.T) {
	_, _, _, stop := startTestTracer(t)
	defer stop()

	child, ctx := StartSpanFromContext(nil, "http.request")
	assert := assert.New(t)
	// ensure the returned context works
	assert.Nil(ctx.Value("not_found_key"))

	internalSpan, ok := child.(*span)
	assert.True(ok)
	assert.Equal("http.request", internalSpan.Name)

	// the returned context includes the span
	ctxSpan, ok := SpanFromContext(ctx)
	assert.True(ok)
	assert.Equal(child, ctxSpan)
}
