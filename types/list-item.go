package types

import (
	"believer/movies/utils"
	"fmt"
)

type List struct {
	Description string `db:"description"`
	ID          string `db:"id"`
	Name        string `db:"name"`
	Rank        int    `db:"rank"`
	Slug        string `db:"slug"`
	Source      string `db:"source"`
}

func (l List) Title() string {
	return l.Name
}

func (l List) Subtitle() string {
	return l.Source
}

func (l List) Href() string {
	return utils.CreateSelfHealingUrl("list", l.Slug, l.ID)
}

type ListItem struct {
	Name     string `db:"name"`
	LinkName string `db:"link_name"`
	ID       string `db:"id"`
	Count    int    `db:"count"`
}

func (l ListItem) LinkTo(root string) string {
	if l.LinkName != "" {
		return fmt.Sprintf("/%s/%s-%s", root, utils.Slugify(l.LinkName), l.ID)
	}
	return fmt.Sprintf("/%s/%s-%s", root, utils.Slugify(l.Name), l.ID)
}

func (l ListItem) FormattedCount() string {
	return utils.Formatter().Sprintf("%d", l.Count)
}
