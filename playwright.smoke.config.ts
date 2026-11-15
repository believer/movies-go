import { defineConfig, devices } from "@playwright/test"

// Production smoke check, run after each deploy with make smoke. It has its
// own config so it never starts the local server or rebuilds the E2E
// database, and lives outside tests/ so a normal E2E run never selects it.
export default defineConfig({
	testDir: "./smoke",
	forbidOnly: !!process.env.CI,
	// A deploy can still be rolling out machines, so retry before failing
	retries: 2,
	reporter: process.env.CI ? "html" : [["list"], ["html", { open: "never" }]],
	use: {
		baseURL: process.env.SMOKE_URL || "https://movies.willcodefor.beer",
		trace: "on-first-retry",
	},
	projects: [
		{
			name: "smoke",
			use: { ...devices["Desktop Chrome"] },
		},
	],
})
