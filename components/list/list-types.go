package list

type DataListItem struct {
	Label string `db:"name"`
	Value string `db:"value"`
}
