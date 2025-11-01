import { expect, Page, test } from '@playwright/test';

const url = process.env.E2E_URL || 'https://movies.willcodefor.beer'

async function search(page: Page, query: string) {
  const searchbox = page.getByRole('searchbox', { name: 'Find a movie, actor, director' })

  await searchbox.click();
  await searchbox.press('ControlOrMeta+a');
  await searchbox.fill(query);
  await searchbox.press('Enter');
}

test('test pages', async ({ page }) => {
  await page.goto(url);

  // Go to moviee
  await page.getByRole('link', { name: 'Fast X' }).click();
  await expect(page.getByRole('heading', { name: 'Metadata' })).toBeVisible();

  // Year
  await page.getByRole('link', { name: '2023-05-17' }).click();
  await expect(page.getByRole('heading', { name: '2023' })).toBeVisible();
  await page.getByRole('link', { name: 'Back' }).click();

  // Series
  await expect(page.getByRole('heading', { name: 'Fast X' })).toBeVisible();
  await page.getByRole('link', { name: 'Fast & Furious #10' }).click();
  await expect(page.getByRole('link', { name: 'The Fast and the Furious', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Back' }).click();

  // Genre
  await page.getByRole('link', { name: 'Adventure' }).click();
  await expect(page.getByRole('heading', { name: 'Adventure' })).toBeVisible();
  await page.getByRole('link', { name: 'Back', exact: true }).click();

  // Language
  await page.getByRole('link', { name: 'English' }).click();
  await expect(page.getByRole('heading', { name: 'English' })).toBeVisible();
  await page.getByRole('link', { name: 'Back' }).click();

  // Rating
  await page.getByRole('link', { name: '8' }).click();
  await expect(page.getByRole('heading', { name: 'Movies rated' })).toBeVisible();
  await page.getByRole('link', { name: 'Back' }).click();

  // Stats
  await page.getByRole('link', { name: 'Stats' }).click();
  await expect(page.getByText('Unique movies seen')).toBeVisible();

  // Watchlist
  await page.getByRole('link', { name: 'Watchlist' }).click();
  await expect(page.getByRole('heading', { name: 'Watchlist' })).toBeVisible();

  // Search
  /// Movie
  await page.getByRole('link', { name: 'Home' }).click();
  await search(page, 'dark knig');
  await expect(page.getByRole('link', { name: 'The Dark Knight', exact: true })).toBeVisible();

  await search(page, 'movie:batman beg');
  await expect(page.getByRole('link', { name: 'Batman Begins', exact: true })).toBeVisible();
  
  /// Cast
  await search(page, 'cast:paul');
  await page.getByRole('link', { name: 'Paul Rudd' }).click();
  await expect(page.getByRole('heading', { name: 'Paul Rudd', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Home' }).click();
  
  /// Director
  await search(page, 'director:steven');
  await expect(page.getByRole('link', { name: 'Steven Spielberg' , exact: true})).toBeVisible();
  
  /// Producer
  await search(page, 'producer:maTT');
  await expect(page.getByRole('link', { name: 'Matthew Vaughn' , exact: true})).toBeVisible();
  
  /// Composer
  await search(page, 'composer:HaNs');
  await expect(page.getByRole('link', { name: 'Hans Zimmer' , exact: true})).toBeVisible();
});
