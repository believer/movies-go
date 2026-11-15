import { acceptDialogs, expect, test } from "./fixtures"

test("a user starts Now playing for a movie in the diary", async ({
	page,
	user: _,
}) => {
	await page.goto("/now-playing/manage")
	await expect(
		page.getByText("No movies currently being watched.")
	).toBeVisible()

	await page.getByLabel("IMDb ID or TMDB ID").fill("tt0107290")
	await page.getByLabel("Position (minutes)").fill("30")
	await page.getByRole("button", { name: "Add Movie" }).click()

	const item = page.getByRole("listitem").filter({ hasText: "Jurassic Park" })

	await expect(
		page.getByRole("heading", { name: "Currently Watching (1)" })
	).toBeVisible()
	await expect(item).toContainText("Progress:")

	// The feed and the movie page show it, even without a Watch
	await page.goto("/")
	await expect(page.getByRole("link", { name: "Now watching" })).toBeVisible()
	await expect(page.getByText("No movies seen")).toBeVisible()
	await page.getByRole("link", { name: "Jurassic Park" }).click()

	await expect(page.getByText("Currently watching")).toBeVisible()
	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched"
	)
})

test("Now playing brings in a movie that isn't in the diary", async ({
	page,
	user: _,
}) => {
	acceptDialogs(page)

	await page.goto("/now-playing/manage")
	await page.getByLabel("IMDb ID or TMDB ID").fill("tt9000003")
	await page.getByLabel("Position (minutes)").fill("44")
	await page.getByRole("button", { name: "Add Movie" }).click()

	const item = page
		.getByRole("listitem")
		.filter({ hasText: "The Pending Reel" })

	await expect(item).toContainText("(50%)")

	await item.getByRole("button", { name: "Edit" }).click()
	await page.getByRole("button", { name: "Delete Watch" }).click()

	await expect(
		page.getByText("No movies currently being watched.")
	).toBeVisible()
})
