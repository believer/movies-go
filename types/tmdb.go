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

type TmdbCompany struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type TmdbCountry struct {
	ID   string `json:"iso_3166_1"`
	Name string `json:"name"`
}

// https://api.themoviedb.org/3/movie/{id}
type MovieDetailsResponse struct {
	Genres              []TmdbGenre    `json:"genres"`
	ImdbId              string         `json:"imdb_id"`
	Overview            string         `json:"overview"`
	Poster              string         `json:"poster_path"`
	ProductionCompanies []TmdbCompany  `json:"production_companies"`
	ProductionCountries []TmdbCountry  `json:"production_countries"`
	ReleaseDate         string         `json:"release_date"`
	Runtime             int            `json:"runtime"`
	SpokenLanguages     []TmdbLanguage `json:"spoken_languages"`
	Tagline             string         `json:"tagline"`
	Title               string         `json:"title"`
	TmdbId              int            `json:"id"`
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

// https://api.themoviedb.org/3/movie/{id}/watch/providers
type MovieWatchProvidersResponse struct {
	Results ProviderCountries `json:"results"`
}

type ProviderCountries struct {
	SE ProviderCountry `json:"SE"`
}

type ProviderCountry struct {
	Ads          []Provider `json:"ads"`
	Buy          []Provider `json:"buy"`
	Free         []Provider `json:"free"`
	Rent         []Provider `json:"rent"`
	Subscription []Provider `json:"flatrate"`
}

type Provider struct {
	Name string `json:"provider_name"`
}
