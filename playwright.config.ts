import { defineConfig, devices } from "@playwright/test"
import { adminSecret, baseURL, databaseURL, port, tmdbBaseURL } from "./e2e/env"

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
	testDir: "./e2e",
	/* The production smoke check has its own config, playwright.smoke.config.ts */
	testIgnore: "smoke/**",
	/* Rebuild the E2E database from the schema and seed once per run */
	globalSetup: "./e2e/global-setup.ts",
	/* Run tests in files in parallel */
	fullyParallel: true,
	/* Fail the build on CI if you accidentally left test.only in the source code. */
	forbidOnly: !!process.env.CI,
	/* Retry on CI only */
	retries: process.env.CI ? 2 : 0,
	/* Half the cores, locally and on CI. Tests that write log in as a fresh
	 * user, so workers don't step on each other's data. */
	workers: "50%",
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

	/* Start the app against the E2E database, rebuilt from the seed by the global
	 * setup, and the TMDB stub it starts. The process env wins over the
	 * checkout's env files, so a .env.local can't point it elsewhere. In
	 * development the login cookie isn't Secure, so it's kept over http.
	 *
	 * The server is built and exec'd, so Playwright stops the server itself
	 * rather than a go run parent that can leave it behind. A server already
	 * on the port fails the run instead of being reused, since it may run old
	 * code or env. Set E2E_REUSE=1 to reuse one on purpose. */
	webServer: {
		command: "go build -o tmp/e2e-server . && exec ./tmp/e2e-server",
		url: `${baseURL}/health`,
		env: {
			PORT: String(port),
			DATABASE_URL: databaseURL,
			ADMIN_SECRET: adminSecret,
			APP_ENV: "development",
			TMDB_BASE_URL: tmdbBaseURL,
			TMDB_API_KEY: "e2e",
		},
		reuseExistingServer: process.env.E2E_REUSE === "1",
		timeout: 120_000,
	},
})
