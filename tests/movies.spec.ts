import { expect, Page, test } from "@playwright/test"

const url = process.env.E2E_URL || "https://movies.willcodefor.beer"

async function search(page: Page, query: string) {
	const searchbox = page.getByRole("searchbox", {
		name: "Find a movie, actor, director",
	})

	await searchbox.click()
	await searchbox.press("ControlOrMeta+a")
	await searchbox.fill(query)
	await searchbox.press("Enter")
}

test("test pages", async ({ page }) => {
	await page.goto(url)

	// Go to first movie
	const firstMovie = page.locator(".feed-list__card").nth(0)
	const title = await firstMovie.getByRole("heading").textContent()

	await firstMovie.click()
	await expect(page.getByRole("heading", { name: "Metadata" })).toBeVisible()

	// Year
	const yearLink = page.getByLabel("Release date")
	const year = ((await yearLink.textContent()) ?? "").substring(0, 4)

	await yearLink.click()
	await page.waitForURL(`**\/year\/${year}`)
	await expect(page.getByRole("heading", { name: year })).toBeVisible()
	await page.getByRole("link", { name: "Back" }).click()

	await expect(page.getByRole("heading", { name: title! })).toBeVisible()

	// Series
	const series = await page.locator("dd#series").count()

	if (series > 0) {
		const seriesLink = page.getByLabel("Series")
		const text = await seriesLink.textContent()
		const clean = text!.replace(/\s#\d{1,}/, "")

		await seriesLink.click()
		await expect(
			page.getByRole("heading", { name: clean, exact: true })
		).toBeVisible()
		await page.getByRole("link", { name: "Back" }).click()
	}

	// Genre
	const firstGenre = await page
		.locator(".contents:has(#genres) > dt > a:first-child")
		.textContent()

	if (firstGenre) {
		await page.getByRole("link", { name: firstGenre }).click()
		await expect(page.getByRole("heading", { name: firstGenre })).toBeVisible()
		await page.getByRole("link", { name: "Back", exact: true }).click()
	}

	// Language
	const firstLanguage = await page
		.locator(".contents:has(#languages) > dt > a:first-child")
		.textContent()

	if (firstLanguage) {
		await page.getByRole("link", { name: "English" }).click()
		await expect(page.getByRole("heading", { name: "English" })).toBeVisible()
		await page.getByRole("link", { name: "Back" }).click()
	}

	// Rating
	await page.getByLabel("rating").click()
	await expect(
		page.getByRole("heading", { name: "Movies rated" })
	).toBeVisible()
	await page.getByRole("link", { name: "Back" }).click()

	// Stats
	await page.getByRole("link", { name: "Stats" }).click()
	await expect(page.getByText("Unique movies seen")).toBeVisible()

	// Watchlist
	await page.getByRole("link", { name: "Watchlist" }).click()
	await expect(page.getByRole("heading", { name: "Watchlist" })).toBeVisible()

	// Search
	/// Movie
	await page.getByRole("link", { name: "Home" }).click()
	await search(page, "dark knig")
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toBeVisible()

	await search(page, "movie:batman beg")
	await expect(
		page.getByRole("link", { name: "Batman Begins", exact: true })
	).toBeVisible()

	/// Cast
	await search(page, "cast:paul")
	await page.getByRole("link", { name: "Paul Rudd" }).click()
	await expect(
		page.getByRole("heading", { name: "Paul Rudd", exact: true })
	).toBeVisible()
	await page.getByRole("link", { name: "Home" }).click()

	/// Director
	await search(page, "director:steven")
	await expect(
		page.getByRole("link", { name: "Steven Spielberg", exact: true })
	).toBeVisible()

	/// Producer
	await search(page, "producer:maTT")
	await expect(
		page.getByRole("link", { name: "Matthew Vaughn", exact: true })
	).toBeVisible()

	/// Composer
	await search(page, "composer:HaNs")
	await page.getByRole("link", { name: "Hans Zimmer", exact: true }).click()
	await expect(page.getByText("Academy Awards")).toBeVisible()
	await expect(page.getByText("Dune Won2021")).toBeVisible()
})
