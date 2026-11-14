import { createHmac, randomUUID } from "node:crypto"
import { test as base } from "@playwright/test"
import pg from "pg"
import { adminSecret, baseURL, databaseURL } from "./env"

export { expect } from "@playwright/test"

export type User = { id: number; username: string; password: string }

async function query<T extends pg.QueryResultRow>(
	text: string,
	values: unknown[]
): Promise<T[]> {
	const client = new pg.Client({ connectionString: databaseURL })

	await client.connect()

	try {
		const { rows } = await client.query<T>(text, values)

		return rows
	} finally {
		await client.end()
	}
}

/* A new user with the seed's password, "password". Write tests each use
 * their own, so their Watches, Ratings and the rest never meet another
 * test's. */
export async function createUser(): Promise<User> {
	const username = `e2e-${randomUUID()}`
	const [{ id }] = await query<{ id: number }>(
		`INSERT INTO "user" (username, name, password_hash)
		SELECT $1, $1, password_hash FROM "user" WHERE id = 1
		RETURNING id`,
		[username]
	)

	return { id, username, password: "password" }
}

/* Logs a Watch without going through the app, for tests of what a Watch
 * unlocks, such as Ratings and Reviews */
export async function logWatch(user: User, movieId: number, date: string) {
	await query(
		"INSERT INTO seen (movie_id, user_id, date) VALUES ($1, $2, $3)",
		[movieId, user.id, date]
	)
}

const base64url = (value: string | Buffer) =>
	Buffer.from(value).toString("base64url")

/* A token as the login form would set it: HS256 over the user's ID, as a
 * string, signed with the server's ADMIN_SECRET */
export function mintToken(userId: number): string {
	const header = base64url(JSON.stringify({ alg: "HS256", typ: "JWT" }))
	const payload = base64url(
		JSON.stringify({
			id: String(userId),
			exp: Math.floor(Date.now() / 1000) + 60 * 60,
		})
	)
	const signature = base64url(
		createHmac("sha256", adminSecret).update(`${header}.${payload}`).digest()
	)

	return `${header}.${payload}.${signature}`
}

/* The user fixture logs the page in as a fresh user, without the login
 * form */
export const test = base.extend<{ user: User }>({
	user: async ({ context }, use) => {
		const user = await createUser()

		await context.addCookies([
			{ name: "token", value: mintToken(user.id), url: baseURL },
		])

		await use(user)
	},
})
