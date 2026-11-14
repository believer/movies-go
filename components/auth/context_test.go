package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticated(t *testing.T) {
	assert.False(t, Authenticated(context.Background()), "missing key")
	assert.False(t, Authenticated(context.WithValue(context.Background(), "IsAuthenticated", false)))
	assert.True(t, Authenticated(context.WithValue(context.Background(), "IsAuthenticated", true)))
}
