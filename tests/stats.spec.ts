import { expect, test } from "@playwright/test"
import { section } from "./section"

test("stats count Watches", async ({ page }) => {
	await page.goto("/")
	await page.getByRole("link", { name: "Stats" }).click()

	const stat = (name: string) =>
		page.locator(".contents").filter({ hasText: name }).locator("dt")

	await expect(stat("Unique movies seen")).toHaveText("11")
	await expect(stat("Movies seen with rewatches")).toHaveText("12")
	await expect(stat("Reviews")).toHaveText("2")
})

test("stats show this year's views", async ({ page }) => {
	await page.goto("/stats")

	// These views follow the clock, so only their structure is checked
	const year = String(new Date().getFullYear())

	await expect(
		page.getByRole("link", { name: "Ratings this year" })
	).toHaveAttribute("href", `/stats/ratings/${year}`)
	await expect(
		page.getByRole("heading", { name: "Seen this year by month" })
	).toBeVisible()
})

test("ratings for a past year are grouped by Rating", async ({ page }) => {
	await page.goto("/stats/ratings/2021")

	await expect(
		page.getByRole("heading", { name: "Ratings in 2021" })
	).toBeVisible()

	const ratingGroup = (rating: string) =>
		section(page, rating, true)
			.getByRole("listitem")
			.getByRole("link", { name: /^\D/ })

	// The Dark Knight's later, lower Rating is the one that counts
	await expect(ratingGroup("9")).toHaveText(["The Dark Knight"])
	await expect(ratingGroup("8")).toHaveText(["Dune"])
	await expect(ratingGroup("6")).toHaveText(["Ant-Man"])
})
