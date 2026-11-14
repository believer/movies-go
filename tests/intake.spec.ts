import { expect, test } from "./fixtures"

// Each test brings in its own made-up movie from the TMDB stub's fixtures,
// so parallel runs never write the same movie

test("logging a Watch of a new movie brings it in from TMDB", async ({
	page,
	user: _,
}) => {
	await page.goto("/movie/new")
	await page.getByLabel("IMDb ID or TMDB ID").fill("tt9000001")
	await page.getByLabel("Rating").fill("8")
	await page.getByRole("button", { name: "Add" }).click()

	await expect(
		page.getByRole("heading", { name: "The Stub Lighthouse", level: 1 })
	).toBeVisible()
	await expect(page.getByLabel("Release date")).toHaveText("1961-03-14")
	await expect(page.getByLabel("My rating")).toHaveText("8")
	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched (1)"
	)
	await expect(
		page.getByRole("link", { name: "Documentary", exact: true })
	).toBeVisible()
	await expect(
		page.getByRole("link", { name: "Greta Lindqvist" })
	).toBeVisible()
	await expect(page.getByRole("link", { name: "Bo Fixtur" })).toBeVisible()
	await expect(
		page.getByRole("link", { name: "Stub Pictures", exact: true })
	).toBeVisible()

	// Watch providers come from TMDB too, and this user has picked none
	await expect(page.getByText("Select your providers in")).toBeVisible()
	await page.getByText("Other providers").click()
	await expect(page.getByText("Stubflix")).toBeVisible()
})

test("a movie found with TMDB search is brought in by its TMDB ID", async ({
	page,
	user: _,
}) => {
	await page.goto("/movie/new")
	await page.getByLabel("Search").pressSequentially("fixture garden")
	await page.getByRole("link", { name: "The Fixture Garden (1962)" }).click()

	await expect(page.getByLabel("IMDb ID or TMDB ID")).toHaveValue("9000002")

	await page.getByLabel("Rating").fill("6")
	await page.getByRole("button", { name: "Add" }).click()

	await expect(
		page.getByRole("heading", { name: "The Fixture Garden", level: 1 })
	).toBeVisible()
	await expect(page.getByLabel("My rating")).toHaveText("6")
	await expect(page.getByRole("link", { name: "Sigrid Holm" })).toBeVisible()
})

test("a movie TMDB doesn't know isn't added", async ({ page, user: _ }) => {
	await page.goto("/movie/new")
	await page.getByLabel("IMDb ID or TMDB ID").fill("tt9999999")
	await page.getByLabel("Rating").fill("5")
	await page.getByRole("button", { name: "Add" }).click()

	await expect(page.getByText("Movie not found on TMDB")).toBeVisible()
	await expect(page).toHaveURL(/\/movie\/new$/)
})
