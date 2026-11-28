package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlural(t *testing.T) {
	assert.Equal(t, "0 nominations", Plural(Nominations, 0))
	assert.Equal(t, "1 nomination", Plural(Nominations, 1))
	assert.Equal(t, "2 nominations", Plural(Nominations, 2))
	assert.Equal(t, "1 win", Plural(Wins, 1))
	assert.Equal(t, "3 wins", Plural(Wins, 3))
	assert.Equal(t, "1 user", Plural(Users, 1))
	assert.Equal(t, "5 users", Plural(Users, 5))
	assert.Equal(t, "1 day", Plural(Days, 1))
	assert.Equal(t, "49 days", Plural(Days, 49))
	assert.Equal(t, "1 movie", Plural(Movies, 1))
	assert.Equal(t, "3 movies", Plural(Movies, 3))
}
