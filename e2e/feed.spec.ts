import { expect, test } from "@playwright/test"

test("the feed lists the latest Watch first", async ({ page }) => {
	await page.goto("/")

	// biome-ignore lint/plugin/no-markup-locator: a feed card is an unlabelled <li>
	const firstCard = page.locator(".feed-list__card").first()

	await expect(firstCard.getByRole("heading")).toHaveText(
		"Kingsman: The Secret Service"
	)
	// biome-ignore lint/plugin/no-markup-locator: the <time> title attribute is what the test checks
	await expect(firstCard.locator("time")).toHaveText("April 9 2022")
})

test("the feed groups Watches by year", async ({ page }) => {
	await page.goto("/")

	for (const year of ["2022", "2021", "2020", "2019"]) {
		await expect(
			page.getByRole("link", { name: year, exact: true })
		).toBeVisible()
	}
})

test("a feed card opens its movie", async ({ page }) => {
	await page.goto("/")
	await page.getByRole("link", { name: "Parasite" }).click()

	await expect(
		page.getByRole("heading", { name: "Parasite", level: 1 })
	).toBeVisible()
	await expect(page.getByText("Ki-taek's family")).toBeVisible()
})

test("the feed lists a movie with no overview", async ({ page }) => {
	await page.goto("/")

	await expect(page.getByRole("link", { name: "Jurassic Park" })).toBeVisible()
})
