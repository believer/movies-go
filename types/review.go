package types

import (
	"fmt"
)

type Review struct {
	ID      int    `db:"id"`
	Content string `db:"content"`
	Private bool   `db:"private"`
}

func (r *Review) Edit(id int) string {
	return fmt.Sprintf("/review/%d/edit?movieId=%d", r.ID, id)
}

func (r *Review) Delete(id int) string {
	return fmt.Sprintf("/review/%d?movieId=%d", r.ID, id)
}
