import type { Page } from "@playwright/test"

export async function search(page: Page, query: string) {
	const searchbox = page.getByRole("searchbox", {
		name: "Find a movie, actor, director",
	})

	await searchbox.fill(query)
	await searchbox.press("Enter")
}
