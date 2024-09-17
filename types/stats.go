package types

type Stats struct {
	UniqueMovies      int `db:"unique_movies"`
	SeenWithRewatches int `db:"seen_with_rewatches"`
	TotalRuntime      int `db:"total_runtime"`
}

type MovieStats struct {
	Title string `db:"title"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

type PersonStats struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

type GraphData struct {
	Label string `db:"label"`
	Value int    `db:"value"`
}

type Bar struct {
	Label     string
	Value     int
	BarHeight int
	BarWidth  int
	BarX      int
	BarY      int
	LabelX    float64
	LabelY    float64
	ValueX    float64
	ValueY    int
}

type Genre struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}
