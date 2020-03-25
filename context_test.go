package instana_test

import (
	"context"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanFromContext_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	span := tracer.StartSpan("test")
	ctx := instana.ContextWithSpan(context.Background(), span)

	sp, ok := instana.SpanFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, span, sp)
}

func TestSpanFromContext_NoActiveSpan(t *testing.T) {
	_, ok := instana.SpanFromContext(context.Background())
	assert.False(t, ok)
}