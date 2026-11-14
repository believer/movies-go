import { expect, test } from "./fixtures"

test("a user logs a Watch of a movie in the diary", async ({
	page,
	user: _,
}) => {
	await page.goto("/movie/batman-begins-1")
	await expect(page.getByText("No watches")).toBeVisible()
	await page.getByRole("link", { name: "Add watch" }).click()

	await expect(page.getByText("Batman Begins", { exact: true })).toBeVisible()
	await expect(page.getByLabel("IMDb ID")).toHaveValue("tt0372784")

	await page.getByLabel("Rating").fill("7")
	await page.getByText("Additional fields").click()
	await page.getByLabel("Watched at").fill("2018-06-01T20:00")
	await page.getByRole("button", { name: "Add" }).click()

	await expect(
		page.getByRole("heading", { name: "Batman Begins", level: 1 })
	).toBeVisible()
	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched (1)"
	)
	await expect(page.getByText("June 01, 2018")).toBeVisible()
	await expect(page.getByLabel("My rating")).toHaveText("7")

	// The Watch leads the user's own feed
	await page.goto("/")
	const firstCard = page.locator(".feed-list__card").first()
	await expect(firstCard.getByRole("heading")).toHaveText("Batman Begins")
	await expect(page.locator(".feed-list__card")).toHaveCount(1)
})

test("a user adds and deletes a rewatch", async ({ page, user: _ }) => {
	page.on("dialog", (dialog) => dialog.accept())

	await page.goto("/movie/new?imdbId=tt0816692&id=4")
	await page.getByLabel("Rating").fill("8")
	await page.getByText("Additional fields").click()
	await page.getByLabel("Watched at").fill("2017-02-03T19:00")
	await page.getByRole("button", { name: "Add" }).click()

	const watched = page.getByRole("heading", { name: /^Watched/ })

	await expect(watched).toHaveText("Watched (1)")
	await page.getByRole("link", { name: "Add new watch" }).click()
	await expect(watched).toHaveText("Watched (2)")

	// The rewatch is logged now, so delete the 2017 Watch and check it's gone
	const first = page
		.getByRole("listitem")
		.filter({ hasText: "February 03, 2017" })
	await first.getByRole("button", { name: "Delete" }).click()

	await expect(watched).toHaveText("Watched (1)")
	await expect(page.getByText("February 03, 2017")).toHaveCount(0)
})
