import { readFile } from "node:fs/promises"
import { createServer, type Server } from "node:http"
import path from "node:path"

const fixtures = path.join(__dirname, "fixtures")

/* Serves TMDB responses from fixtures/, keyed by request path: a request
 * for /3/movie/tt9000001/credits gets fixtures/movie/tt9000001/credits.json.
 * Query strings are ignored, so a search returns the same results for any
 * query. A path without a fixture gets a 404, as TMDB gives for an unknown
 * movie. */
export function startTmdbStub(port: number): Promise<Server> {
	const server = createServer(async (req, res) => {
		const { pathname } = new URL(req.url ?? "/", "http://localhost")
		const key = path.normalize(pathname.replace(/^\/3\//, ""))
		const file = path.join(fixtures, `${key}.json`)

		try {
			if (!file.startsWith(fixtures + path.sep)) {
				throw new Error("outside fixtures")
			}

			const body = await readFile(file)

			res.writeHead(200, { "Content-Type": "application/json" })
			res.end(body)
		} catch {
			res.writeHead(404, { "Content-Type": "application/json" })
			res.end(JSON.stringify({ success: false, status_code: 34 }))
		}
	})

	return new Promise((resolve, reject) => {
		server.once("error", reject)
		server.listen(port, () => resolve(server))
	})
}
