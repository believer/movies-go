package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticated(t *testing.T) {
	// A string key, since the middleware sets it with fiber's Locals and
	// fasthttp's context reads Locals back by string
	assert.False(t, Authenticated(context.Background()), "missing key")
	assert.False(t, Authenticated(context.WithValue(context.Background(), "IsAuthenticated", false))) //nolint:staticcheck
	assert.True(t, Authenticated(context.WithValue(context.Background(), "IsAuthenticated", true)))   //nolint:staticcheck
}
