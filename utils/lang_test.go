package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluralMessage(t *testing.T) {
	assert.Equal(t, "0 nominations", PluralMessage(NominationKey, 0))
	assert.Equal(t, "1 nomination", PluralMessage(NominationKey, 1))
	assert.Equal(t, "2 nominations", PluralMessage(NominationKey, 2))
	assert.Equal(t, "1 win", PluralMessage(WinKey, 1))
	assert.Equal(t, "3 wins", PluralMessage(WinKey, 3))
	assert.Equal(t, "1 user", PluralMessage(UsersKey, 1))
	assert.Equal(t, "5 users", PluralMessage(UsersKey, 5))
}
