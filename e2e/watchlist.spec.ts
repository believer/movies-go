import { expect, test } from "./fixtures"

test("the Watchlist lists the user's movies", async ({ page }) => {
	await page.goto("/")
	await page.getByRole("link", { name: "Watchlist" }).click()

	await expect(page.getByRole("heading", { name: "Watchlist" })).toBeVisible()
	await expect(page.getByRole("heading", { name: "Movies (1)" })).toBeVisible()
	await expect(page.getByRole("link", { name: "The Lion King" })).toBeVisible()
})

test("a user adds a movie to the Watchlist, and a Watch removes it", async ({
	page,
	user: _,
}) => {
	await page.goto("/movie/new")
	await page.getByLabel("IMDb ID or TMDB ID").fill("tt0211915")
	await page.getByLabel("Add to watchlist").check()
	await page.getByRole("button", { name: "Add" }).click()

	await expect(
		page.getByRole("heading", { name: "Amélie", level: 1 })
	).toBeVisible()
	await expect(page.getByRole("link", { name: "In watchlist" })).toBeVisible()

	await page.goto("/watchlist")
	await expect(page.getByRole("heading", { name: "Movies (1)" })).toBeVisible()
	await expect(page.getByRole("link", { name: "Amélie" })).toBeVisible()
	// User 1's Watchlist doesn't leak in
	await expect(page.getByRole("link", { name: "The Lion King" })).toHaveCount(0)

	await page.getByRole("link", { name: "Amélie" }).click()
	await page.getByRole("link", { name: "Add watch" }).click()
	await page.getByLabel("Rating").fill("9")
	await page.getByRole("button", { name: "Add" }).click()

	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched (1)"
	)

	await page.goto("/watchlist")
	await expect(page.getByRole("link", { name: "Amélie" })).toHaveCount(0)
})
