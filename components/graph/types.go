package graph

type GraphData struct {
	Label string `db:"label" json:"label"`
	Value int    `db:"value" json:"value"`
}

type Bar struct {
	Label     string  `json:"label"`
	Value     int     `json:"value"`
	BarHeight int     `json:"barHeight"`
	BarWidth  int     `json:"barWidth"`
	BarX      int     `json:"barX"`
	BarY      int     `json:"barY"`
	LabelX    float64 `json:"labelX"`
	LabelY    float64 `json:"labelY"`
	ValueX    float64 `json:"valueX"`
	ValueY    int     `json:"valueY"`
}
