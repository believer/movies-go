import { expect, test } from "@playwright/test"
import { section } from "./section"

/* Screenshots of the sections a markup or CSS change is most likely to shift.
 * The committed baselines are Linux Chromium from CI. Fonts render
 * differently elsewhere, so locally they run only with VISUAL=1 and write
 * baselines of their own, which git ignores. */
test.skip(
	({ browserName }) =>
		browserName !== "chromium" || !(process.env.CI || process.env.VISUAL),
	"Visual baselines are Chromium only, on CI or with VISUAL=1"
)

/* A missing baseline is written and fails the test. A retry would then pass
 * against it, so retries would hide that the baseline is new. */
test.describe.configure({ retries: 0 })

const viewports = {
	desktop: { width: 1280, height: 720 },
	// Under the description list's 400px container query
	narrow: { width: 375, height: 800 },
}

for (const [name, viewport] of Object.entries(viewports)) {
	test.describe(name, () => {
		test.use({ viewport })

		for (const movie of ["the-dark-knight-2", "parasite-11"]) {
			test(`${movie} metadata`, async ({ page }) => {
				await page.goto(`/movie/${movie}`)

				await expect(section(page, "Metadata")).toHaveScreenshot(
					`${movie}-metadata-${name}.png`
				)
			})
		}

		test("stats", async ({ page }) => {
			await page.goto("/stats")

			for (const title of ["Stats", "Awards"]) {
				await expect(section(page, title, true)).toHaveScreenshot(
					`stats-${title.toLowerCase()}-${name}.png`
				)
			}
		})
	})
}
