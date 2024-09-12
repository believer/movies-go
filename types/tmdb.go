package types

type SearchResult struct {
	Title string `json:"title"`
	Id    int    `json:"id"`
}

// https://api.themoviedb.org/3/search/movie
type SearchMovieResponse struct {
	Results []SearchResult `json:"results"`
}

type TmdbGenre struct {
	Name string `json:"name"`
}

// https://api.themoviedb.org/3/movie/{id}
type MovieDetailsResponse struct {
	Title       string      `json:"title"`
	Runtime     int         `json:"runtime"`
	ReleaseDate string      `json:"release_date"`
	ImdbId      string      `json:"imdb_id"`
	Overview    string      `json:"overview"`
	Poster      string      `json:"poster_path"`
	Tagline     string      `json:"tagline"`
	Genres      []TmdbGenre `json:"genres"`
}

type CastResult struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Character   *string `json:"character"`
	Popularity  float64 `json:"popularity"`
	ProfilePath *string `json:"profile_path"`
}

type CrewResult struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Department  string  `json:"department"`
	Job         *string `json:"job"`
	Popularity  float64 `json:"popularity"`
	ProfilePath *string `json:"profile_path"`
}

// https://api.themoviedb.org/3/movie/{id}/credits
type MovieCreditsResponse struct {
	Cast []CastResult `json:"cast"`
	Crew []CrewResult `json:"crew"`
}
