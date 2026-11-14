import { expect, logWatch, test } from "./fixtures"
import { inception } from "./seed"

test("a user rates a watched movie and changes the Rating", async ({
	page,
	user,
}) => {
	await logWatch(user, inception, "2018-05-04 20:00")

	await page.goto(inception.path)
	await page.getByRole("button", { name: "Add rating" }).click()
	await page.getByRole("spinbutton").fill("4")
	await page.getByRole("button", { name: "Save rating" }).click()

	const rating = page.getByLabel("My rating")

	await expect(rating).toHaveText("4")

	await page.getByRole("button", { name: "Edit", exact: true }).click()
	await expect(page.getByRole("spinbutton")).toHaveValue("4")
	await page.getByRole("spinbutton").fill("2")
	await page.getByRole("button", { name: "Update rating" }).click()
	await expect(rating).toHaveText("2")

	await page.reload()
	await expect(rating).toHaveText("2")

	// The Rating links to the user's own movies with that Rating
	await rating.click()
	await expect(
		page.getByRole("heading", { name: "Movies rated 2" })
	).toBeVisible()
	await expect(page.getByRole("link", { name: "Inception" })).toBeVisible()
})

test("a movie without a Watch can't be rated", async ({ page, user: _ }) => {
	await page.goto(inception.path)

	await expect(page.getByText("No watches")).toBeVisible()
	await expect(page.getByRole("button", { name: "Add rating" })).toHaveCount(0)
})
