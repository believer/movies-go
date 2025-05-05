package types

type OthersStats struct {
	Seen          int     `db:"seen_count"`
	AverageRating float64 `db:"avg_rating"`
}
