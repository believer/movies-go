package types

import (
	"believer/movies/utils"
	"fmt"

	"github.com/a-h/templ"
)

type ListItem struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

func (l ListItem) LinkTo(root string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/%s/%s-%s", root, utils.Slugify(l.Name), l.ID))
}

func (l ListItem) FormattedCount() string {
	return utils.Formatter().Sprintf("%d", l.Count)
}
