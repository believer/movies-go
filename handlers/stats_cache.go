package handlers

import (
	"believer/movies/views"
	"sync"
	"time"
)

type cachedStats struct {
	props     views.StatsProps
	createdAt time.Time
}

var (
	statsCache      = make(map[string]cachedStats)
	statsCacheMutex sync.RWMutex
	cacheTTL        = 10 * time.Minute
)

// InvalidateStatsCache clears the cached stats for a specific user.
func InvalidateStatsCache(userID string) {
	statsCacheMutex.Lock()
	delete(statsCache, userID)
	statsCacheMutex.Unlock()
}
