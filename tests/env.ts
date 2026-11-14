/* Settings shared by the Playwright config, its global setup and the test
 * helpers. The E2E server gets its own port, so it runs beside make dev
 * (8080 and 7331), and the TMDB stub the next one. Override them with
 * E2E_PORT and E2E_TMDB_PORT to run beside another checkout's E2E server. */
export const port = Number(process.env.E2E_PORT || 8090)
export const baseURL = `http://localhost:${port}`

export const tmdbPort = Number(process.env.E2E_TMDB_PORT || 8091)
export const tmdbBaseURL = `http://localhost:${tmdbPort}/3`

export const databaseURL =
	process.env.E2E_DATABASE_URL ||
	"postgres://admin:password@localhost:5433/movies_e2e?sslmode=disable"

/* The server signs and checks tokens with this, so tests can mint their own */
export const adminSecret = "e2e-secret"
