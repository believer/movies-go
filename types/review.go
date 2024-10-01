package types

type Review struct {
	ID      int    `db:"id"`
	Content string `db:"content"`
	Private bool   `db:"private"`
}
