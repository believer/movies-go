import { defineConfig, devices } from "@playwright/test"
import {
	adminSecret,
	baseURL,
	databaseURL,
	port,
	tmdbBaseURL,
} from "./tests/env"

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
	testDir: "./tests",
	/* Rebuild movies_e2e from the schema and seed once per run */
	globalSetup: "./tests/global-setup.ts",
	/* Run tests in files in parallel */
	fullyParallel: true,
	/* Fail the build on CI if you accidentally left test.only in the source code. */
	forbidOnly: !!process.env.CI,
	/* Retry on CI only */
	retries: process.env.CI ? 2 : 0,
	/* Opt out of parallel tests on CI. */
	workers: process.env.CI ? 1 : undefined,
	/* Reporter to use. See https://playwright.dev/docs/test-reporters
	 * Locally the HTML report is written but never opened: opening it serves
	 * the report on failure and blocks until killed. */
	reporter: process.env.CI ? "html" : [["list"], ["html", { open: "never" }]],
	/* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
	use: {
		/* Base URL to use in actions like `await page.goto('')`. */
		baseURL,

		/* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
		trace: "on-first-retry",
	},

	/* Configure projects for major browsers */
	projects: [
		{
			name: "chromium",
			use: { ...devices["Desktop Chrome"] },
		},

		{
			name: "firefox",
			use: { ...devices["Desktop Firefox"] },
		},

		{
			name: "webkit",
			use: { ...devices["Desktop Safari"] },
		},

		/* Test against mobile viewports. */
		// {
		//   name: 'Mobile Chrome',
		//   use: { ...devices['Pixel 5'] },
		// },
		// {
		//   name: 'Mobile Safari',
		//   use: { ...devices['iPhone 12'] },
		// },

		/* Test against branded browsers. */
		// {
		//   name: 'Microsoft Edge',
		//   use: { ...devices['Desktop Edge'], channel: 'msedge' },
		// },
		// {
		//   name: 'Google Chrome',
		//   use: { ...devices['Desktop Chrome'], channel: 'chrome' },
		// },
	],

	/* Start the app against movies_e2e, rebuilt from the seed by the global
	 * setup, and the TMDB stub it starts. The process env wins over the
	 * checkout's env files, so a .env.local can't point it elsewhere. In
	 * development the login cookie isn't Secure, so it's kept over http. */
	webServer: {
		command: "go run .",
		url: `${baseURL}/health`,
		env: {
			PORT: String(port),
			DATABASE_URL: databaseURL,
			ADMIN_SECRET: adminSecret,
			APP_ENV: "development",
			TMDB_BASE_URL: tmdbBaseURL,
			TMDB_API_KEY: "e2e",
		},
		reuseExistingServer: !process.env.CI,
		timeout: 120_000,
	},
})
