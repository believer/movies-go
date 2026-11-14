import { test as base, expect } from "@playwright/test"

// Production data changes with every watch, so these checks only assert that
// pages render, never which movies they show.

// Fails every test on an error response, including the htmx partials a page
// loads after it renders, or an uncaught script error.
const test = base.extend<{ noErrors: void }>({
	noErrors: [
		async ({ page }, use) => {
			const errors: string[] = []

			page.on("response", (response) => {
				if (response.status() >= 400) {
					errors.push(`${response.status()} ${response.url()}`)
				}
			})
			page.on("pageerror", (error) => errors.push(error.message))

			await use()

			await page.waitForLoadState("networkidle")
			expect(errors).toEqual([])
		},
		{ auto: true },
	],
})

test("feed and a movie from it", async ({ page }) => {
	await page.goto("/")

	const firstMovie = page.locator(".feed-list__card").first()
	const title = await firstMovie.getByRole("heading").textContent()

	await firstMovie.click()

	await expect(page.getByRole("heading", { name: title! })).toBeVisible()
	await expect(page.getByRole("heading", { name: "Metadata" })).toBeVisible()
})

test("stats", async ({ page }) => {
	await page.goto("/stats")

	await expect(page.getByRole("heading", { name: "Stats" })).toBeVisible()
	await expect(
		page.getByRole("heading", { name: "Most watched movies" })
	).toBeVisible()
})

test("watchlist", async ({ page }) => {
	await page.goto("/watchlist")

	await expect(
		page.getByRole("heading", { name: "Watchlist", level: 1 })
	).toBeVisible()
	// The heading ends in a count, e.g. Movies (13)
	await expect(page.getByRole("heading", { name: /^Movies/ })).toBeVisible()
})
