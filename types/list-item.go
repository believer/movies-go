package types

import (
	"believer/movies/utils"
	"fmt"
)

type List struct {
	Description string  `db:"description"`
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	Rank        int     `db:"rank"`
	Slug        string  `db:"slug"`
	Source      string  `db:"source"`
	PercentSeen float64 `db:"percent_seen"`
	SeenMovies  int     `db:"seen_movies"`
	TotalMovies int     `db:"total_movies"`
}

func (l List) Title() string {
	return l.Name
}

func (l List) Subtitle() string {
	return l.Source
}

func (l List) Trailing() string {
	return fmt.Sprintf("%.0f%% (%d/%d)", l.PercentSeen, l.SeenMovies, l.TotalMovies)
}

func (l List) Href() string {
	return utils.CreateSelfHealingUrl("list", l.Slug, l.ID)
}

type ListItem struct {
	Name     string `db:"name" json:"name"`
	LinkName string `db:"link_name" json:"linkName"`
	ID       string `db:"id" json:"id"`
	Count    int    `db:"count" json:"count"`
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
