import { expect, test } from "@playwright/test"

test("the release date links to the year page", async ({ page }) => {
	await page.goto("/movie/the-dark-knight-2")

	const releaseDate = page.getByLabel("Release date")

	await expect(releaseDate).toHaveText("2008-07-16")
	await releaseDate.click()

	await expect(page).toHaveURL(/\/year\/2008$/)
	await expect(page.getByRole("heading", { name: "2008" })).toBeVisible()
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toBeVisible()
})

test("the Series links to the Series page", async ({ page }) => {
	await page.goto("/movie/the-dark-knight-2")

	const series = page.getByLabel("Series")

	await expect(series).toHaveText("The Dark Knight Trilogy #2")
	await series.click()

	await expect(
		page.getByRole("heading", { name: "The Dark Knight Trilogy" })
	).toBeVisible()
	await expect(
		page.getByRole("link", { name: "Batman Begins", exact: true })
	).toBeVisible()
	await expect(
		page.getByRole("link", { name: "The Dark Knight", exact: true })
	).toBeVisible()
})

test("a genre links to the genre page", async ({ page }) => {
	await page.goto("/movie/batman-begins-1")
	await page.getByRole("link", { name: "Crime", exact: true }).click()

	await expect(page.getByRole("heading", { name: "Crime" })).toBeVisible()
	await expect(
		page.getByRole("link", { name: "Kingsman: The Secret Service" })
	).toBeVisible()
	await expect(
		page.getByRole("link", { name: "Batman Begins", exact: true })
	).toBeVisible()
})

test("a language links to the language page", async ({ page }) => {
	await page.goto("/movie/amelie-9")

	const language = page.locator("#languages").getByRole("link").first()
	const name = (await language.textContent())?.trim() ?? ""

	expect(name).toBe("French")
	await page.getByRole("link", { name, exact: true }).click()

	// The language page is titled with the language's own name
	await expect(page.getByRole("heading", { name: "Français" })).toBeVisible()
	await expect(page.getByRole("link", { name: "Amélie" })).toBeVisible()
	await expect(page.getByRole("link", { name: "Inception" })).toBeVisible()
	await expect(page.getByRole("link", { name: "Parasite" })).toHaveCount(0)
})

test("the Rating links to movies with the same Rating", async ({ page }) => {
	await page.goto("/movie/inception-3")

	const rating = page.getByLabel("My rating")

	await expect(rating).toHaveText("9")
	await rating.click()

	await expect(
		page.getByRole("heading", { name: "Movies rated 9" })
	).toBeVisible()

	for (const title of ["Inception", "Interstellar", "Parasite"]) {
		await expect(page.getByRole("link", { name: title })).toBeVisible()
	}

	await expect(page.getByRole("link", { name: "Amélie" })).toHaveCount(0)
})
