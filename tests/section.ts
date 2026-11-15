import type { Page } from "@playwright/test"

/* The section headed by this title, not one that merely holds a heading
 * with it further down */
export function section(page: Page, title: string, exact = false) {
	// biome-ignore lint/plugin/no-markup-locator: a <section> without an accessible name has no region role
	return page.locator("section").filter({
		has: page
			// biome-ignore lint/plugin/no-markup-locator: the title must be the section's own child
			.locator(":scope > .section__title")
			.getByRole("heading", { name: title, exact }),
	})
}
