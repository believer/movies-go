import { expect, test } from "@playwright/test"

// User 2 watched and rated The Lion King, and user 1 never did. The app
// falls back to user 1 without a token.
test("another user's Watches don't show", async ({ page }) => {
	await page.goto("/")
	await expect(page.locator(".feed-list__card").first()).toBeVisible()
	await expect(page.getByRole("link", { name: "The Lion King" })).toHaveCount(0)

	await page.goto("/stats/ratings/2021")
	await expect(
		page.getByRole("heading", { name: "Ratings in 2021" })
	).toBeVisible()
	await expect(page.getByRole("link", { name: "The Lion King" })).toHaveCount(0)

	// User 2's own Watch of The Dark Knight isn't listed, only the one they
	// shared with user 1
	await page.goto("/movie/the-dark-knight-2")
	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched (2)"
	)
})
