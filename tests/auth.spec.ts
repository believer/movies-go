import { createUser, expect, test } from "./fixtures"

test("a user logs in with the login form", async ({ page }) => {
	const user = await createUser()

	await page.goto("/login")
	await page.getByLabel("Username").fill(user.username)
	await page.getByLabel("Password").fill("wrong password")
	await page.getByRole("button", { name: "Login" }).click()

	await expect(page.getByText("Invalid username or password")).toBeVisible()

	await page.getByLabel("Password").fill(user.password)
	await page.getByRole("button", { name: "Login" }).click()

	await expect(page).toHaveURL(/\/$/)
	await expect(page.getByRole("link", { name: "Settings" })).toBeVisible()
	await expect(page.getByRole("link", { name: "Login" })).toHaveCount(0)
})

test("a minted token logs in as a fresh user", async ({ page, user }) => {
	await page.goto(`/movie/batman-begins-1`)

	await expect(page.getByRole("link", { name: "Settings" })).toBeVisible()
	await expect(page.getByRole("heading", { name: /^Watched/ })).toHaveText(
		"Watched"
	)
	await expect(page.getByText("No watches")).toBeVisible()
	expect(user.id).toBeGreaterThan(2)
})
