package types

type OthersStats struct {
	Seen          int     `db:"seen_count" json:"seenByUsers"`
	AverageRating float64 `db:"avg_rating" json:"averageRating"`
}
