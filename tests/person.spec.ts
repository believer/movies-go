import { expect, test } from "@playwright/test"

test("a person page lists credits and awards", async ({ page }) => {
	await page.goto("/person/hans-zimmer-5")

	await expect(
		page.getByRole("heading", { name: "Hans Zimmer", level: 1 })
	).toBeVisible()
	await expect(page.getByText("6 credits")).toBeVisible()

	const awards = (name: string) =>
		page.locator("section").filter({
			has: page
				.locator(":scope > .section__title")
				.getByRole("heading", { name }),
		})
	const academyAwards = awards("Academy Awards")

	await expect(
		page.getByRole("heading", { name: "Academy Awards", level: 2 })
	).toHaveText("Academy Awards (3 nominations / 1 win)")
	await expect(
		academyAwards.getByRole("listitem").filter({ hasText: "Dune" })
	).toHaveText(/Won\s*2021/)
	await expect(
		academyAwards.getByRole("listitem").filter({ hasText: "Interstellar" })
	).toHaveText(/Nominated\s*2014/)

	const bafta = awards("BAFTA")

	await expect(
		bafta.getByRole("listitem").filter({ hasText: "Dune" })
	).toHaveText(/Won\s*2022/)
})
