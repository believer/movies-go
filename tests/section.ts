import type { Page } from "@playwright/test"

/* The section headed by this title, not one that merely holds a heading
 * with it further down */
export function section(page: Page, title: string, exact = false) {
	return page.locator("section").filter({
		has: page
			.locator(":scope > .section__title")
			.getByRole("heading", { name: title, exact }),
	})
}
