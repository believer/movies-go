/* Movies from db/seed.sql that write tests set up data for by ID */
export type SeedMovie = { id: number; imdbId: string; path: string }

export const batmanBegins: SeedMovie = {
	id: 1,
	imdbId: "tt0372784",
	path: "/movie/batman-begins-1",
}

export const inception: SeedMovie = {
	id: 3,
	imdbId: "tt1375666",
	path: "/movie/inception-3",
}

export const interstellar: SeedMovie = {
	id: 4,
	imdbId: "tt0816692",
	path: "/movie/interstellar-4",
}

export const antMan: SeedMovie = {
	id: 7,
	imdbId: "tt0478970",
	path: "/movie/ant-man-7",
}
