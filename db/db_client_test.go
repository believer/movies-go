package db

import (
	"believer/movies/db/dbtest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeConnection(t *testing.T) {
	t.Run("caps the open connections", func(t *testing.T) {
		// dbtest skips with -short and fails without TEST_DATABASE_URL
		dbtest.Open(t)
		t.Setenv("DATABASE_URL", os.Getenv("TEST_DATABASE_URL"))

		previous := Client
		t.Cleanup(func() { Client = previous })

		require.NoError(t, InitializeConnection())
		t.Cleanup(CloseConnection)

		assert.Equal(t, maxOpenConns, Client.Stats().MaxOpenConnections)
	})
}
