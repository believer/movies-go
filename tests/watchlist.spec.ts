import { expect, test } from "@playwright/test"

test("the Watchlist lists the user's movies", async ({ page }) => {
	await page.goto("/")
	await page.getByRole("link", { name: "Watchlist" }).click()

	await expect(page.getByRole("heading", { name: "Watchlist" })).toBeVisible()
	await expect(page.getByRole("heading", { name: "Movies (1)" })).toBeVisible()
	await expect(page.getByRole("link", { name: "The Lion King" })).toBeVisible()
})
