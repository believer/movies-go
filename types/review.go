package types

import (
	"fmt"
)

type Review struct {
	ID      int    `db:"id"`
	Content string `db:"content"`
	Private bool   `db:"private"`
}

func (r *Review) Edit() string {
	return fmt.Sprintf("/review/%d/edit", r.ID)
}
