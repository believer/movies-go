import { execSync } from "node:child_process"
import { port, tmdbPort } from "./env"
import { startTmdbStub } from "./tmdb/stub"

/* Rebuild this port's database so every run starts from the seed. CI points make at
 * its Postgres service by setting PSQL (and PGPASSWORD) in the env.
 *
 * The TMDB stub runs in this process for the whole run, and the server is
 * pointed at it, so no test reaches the real TMDB. */
export default async function globalSetup() {
	const stub = await startTmdbStub(tmdbPort)

	execSync("make e2e-db", {
		stdio: "inherit",
		env: { ...process.env, E2E_PORT: String(port) },
	})

	return () => new Promise<void>((resolve) => stub.close(() => resolve()))
}
