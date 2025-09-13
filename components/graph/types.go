package graph

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
