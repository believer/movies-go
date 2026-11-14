import { expect, logWatch, test } from "./fixtures"

test("a user reviews a watched movie and edits the Review", async ({
	page,
	user,
}) => {
	await logWatch(user, 7, "2018-08-10 20:00")

	await page.goto("/movie/ant-man-7")
	await page.getByRole("button", { name: "Add review" }).click()
	await page.getByLabel("Content").fill("Small hero, big heist.")
	await page.getByRole("button", { name: "Add review" }).click()

	await expect(page.getByText("Small hero, big heist.")).toBeVisible()
	await expect(page.getByLabel("Content")).toHaveCount(0)

	await page.reload()
	await expect(page.getByText("Small hero, big heist.")).toBeVisible()

	await page.getByRole("button", { name: "Edit review" }).click()
	await page.getByLabel("Content").fill("Smaller hero, bigger heist.")
	await page.getByRole("button", { name: "Update review" }).click()

	await expect(page.getByText("Smaller hero, bigger heist.")).toBeVisible()
	await page.reload()
	await expect(page.getByText("Smaller hero, bigger heist.")).toBeVisible()
})
