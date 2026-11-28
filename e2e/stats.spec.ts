import { expect, test } from "@playwright/test"
import { section } from "./section"

test("stats count Watches", async ({ page }) => {
	await page.goto("/")
	await page.getByRole("link", { name: "Stats" }).click()

	const stat = (name: string) => page.getByLabel(name, { exact: true })

	await expect(stat("Unique movies seen")).toHaveText("11")
	await expect(stat("Movies seen with rewatches")).toHaveText("12")
	await expect(stat("Reviews")).toHaveText("2")
})

test("stats counts use tabular numbers", async ({ page }) => {
	await page.goto("/stats")

	await expect(
		page.getByLabel("Unique movies seen", { exact: true })
	).toHaveCSS("font-variant-numeric", "tabular-nums")
	await expect(page.getByLabel("Shortest movie", { exact: true })).toHaveCSS(
		"font-variant-numeric",
		"normal"
	)
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

test("hot takes compare the latest Ratings with IMDb's", async ({ page }) => {
	await page.goto("/stats")

	const hotTakes = section(page, "Hot takes", true)
	const titles = (title: string) =>
		hotTakes.getByRole("region", { name: title }).getByRole("link")

	// The Dark Knight's latest Rating, 9, matches IMDb's, so it's in neither
	await expect(titles("Above IMDb")).toHaveText([
		"Spirited Away",
		"Parasite",
		"Interstellar",
		"Inception",
	])
	await expect(titles("Below IMDb")).toHaveText([
		"Ant-Man",
		"Amélie",
		"Kingsman: The Secret Service",
		"Batman Begins",
		"Jurassic Park",
	])
	await expect(
		hotTakes
			.getByRole("region", { name: "Above IMDb" })
			.getByRole("listitem")
			.first()
	).toContainText(/10 vs 8\.6\s*\+1\.4/)
	await expect(hotTakes.getByLabel("Compared to IMDb")).toContainText(
		/0\.1 below IMDb on average\s*11 movies/
	)
})

test("streaks show the records of the days with Watches", async ({ page }) => {
	await page.goto("/stats")

	const streaks = section(page, "Streaks", true)

	await expect(streaks.getByLabel("Longest streak")).toContainText(
		/1 day\s*March 2 2019/
	)
	await expect(streaks.getByLabel("Longest break")).toContainText(
		/279 days\s*Between January 16 2021 and October 23 2021/
	)
	await expect(streaks.getByLabel("Most in one day")).toContainText(
		/1 movie\s*March 2 2019/
	)
	await expect(streaks.getByLabel("Binge days")).toContainText(
		/0 days\s*With more than one movie/
	)
})

test("actor and director pairs need more than one movie", async ({ page }) => {
	await page.goto("/stats")

	// Christopher Nolan directed Bale twice, and the rest of his cast once
	await expect(
		section(page, "Actor & director pairs", true).getByRole("listitem")
	).toHaveText([/^Christian Bale\s*&\s*Christopher Nolan\s*2$/])
})
