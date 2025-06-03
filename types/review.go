package types

import (
	"fmt"

	"github.com/a-h/templ"
)

type Review struct {
	ID      int    `db:"id"`
	Content string `db:"content"`
	Private bool   `db:"private"`
}

func (r *Review) Edit() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/review/%d/edit", r.ID))
}
