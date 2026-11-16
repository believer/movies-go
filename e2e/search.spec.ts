import { expect, test } from "@playwright/test"
import { search } from "./search"

test("search finds movies by title", async ({ page }) => {
	await page.goto("/")

	// The Dark Knight was watched twice, so it's listed twice
	await search(page, "dark knig")
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toHaveCount(2)

	await search(page, "movie:batman beg")
	await expect(
		page.getByRole("link", { name: "Batman Begins", exact: true })
	).toBeVisible()
})

test("search finds movies by their latest Rating", async ({ page }) => {
	await page.goto("/")

	// The Dark Knight was rated 10, then 9
	await search(page, "rating:10")
	await expect(page.getByRole("link", { name: "Spirited Away" })).toBeVisible()
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toHaveCount(0)

	await search(page, "rating:9")
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toHaveCount(1)
})

const people = [
	{ query: "cast:paul", name: "Paul Rudd" },
	{ query: "director:steven", name: "Steven Spielberg" },
	{ query: "producer:maTT", name: "Matthew Vaughn" },
	{ query: "composer:HaNs", name: "Hans Zimmer" },
]

for (const { query, name } of people) {
	test(`search "${query}" finds ${name}`, async ({ page }) => {
		await page.goto("/")
		await search(page, query)
		await page.getByRole("link", { name, exact: true }).click()

		await expect(page.getByRole("heading", { name, exact: true })).toBeVisible()
	})
}
