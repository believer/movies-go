package types

import (
	"believer/movies/utils"
	"fmt"
)

type ListItem struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

func (l ListItem) LinkTo(root string) string {
	return fmt.Sprintf("/%s/%s-%s", root, utils.Slugify(l.Name), l.ID)
}

func (l ListItem) FormattedCount() string {
	return utils.Formatter().Sprintf("%d", l.Count)
}
