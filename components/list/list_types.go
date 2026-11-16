package list

type DataListItem struct {
	Label string `db:"name" json:"label"`
	Value string `db:"value" json:"value"`
}
