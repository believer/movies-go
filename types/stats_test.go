package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHotTakeFormattedDifference(t *testing.T) {
	tests := []struct {
		difference float64
		want       string
	}{
		{3.1, "+3.1"},
		{-6.3, "−6.3"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, HotTake{Difference: tt.difference}.FormattedDifference())
	}
}

func TestIMDbComparisonSummary(t *testing.T) {
	tests := []struct {
		name    string
		average float64
		want    string
	}{
		{"above", 0.42, "0.4 above IMDb on average"},
		{"below", -0.19, "0.2 below IMDb on average"},
		// Rounds to 0.0, which has no side
		{"level", -0.04, "Level with IMDb on average"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IMDbComparison{Movies: 10, AverageDifference: tt.average}.Summary())
		})
	}
}
