package types

type SearchResult struct {
	Overview    string `json:"overview"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Id          int    `json:"id"`
}

// https://api.themoviedb.org/3/search/movie
type SearchMovieResponse struct {
	Results []SearchResult `json:"results"`
}

type TmdbGenre struct {
	Name string `json:"name"`
}

type TmdbLanguage struct {
	ISO639      string `json:"iso_639_1"`
	EnglishName string `json:"english_name"`
	Name        string `json:"name"`
}

// https://api.themoviedb.org/3/movie/{id}
type MovieDetailsResponse struct {
	Genres          []TmdbGenre    `json:"genres"`
	ImdbId          string         `json:"imdb_id"`
	Overview        string         `json:"overview"`
	Poster          string         `json:"poster_path"`
	ReleaseDate     string         `json:"release_date"`
	Runtime         int            `json:"runtime"`
	SpokenLanguages []TmdbLanguage `json:"spoken_languages"`
	Tagline         string         `json:"tagline"`
	Title           string         `json:"title"`
}

type CastResult struct {
	ID          int     `json:"id"`
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
