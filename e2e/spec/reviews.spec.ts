import { test, expect } from '@playwright/test';
import { loginAs } from '../helpers/auth';

// Scenarios from e2e/feat/reviews.feature

const userId = process.env.E2E_TEST_USER_ID;
const albumId = process.env.E2E_TEST_ALBUM_ID;

test('Rating modal opens to the confirmation form', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();

  await expect(page.getByTestId('rating-confirm')).toBeVisible();
  await expect(page.getByTestId('rating-input')).toBeVisible();
  await expect(page.getByTestId('rating-lock-in')).toBeVisible();
});

test('Navigating to the questionnaire from the confirmation form', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();
  await expect(page.getByTestId('rating-confirm')).toBeVisible();

  // The ? button inside the confirm form navigates to the questionnaire
  await page.getByTestId('rating-confirm').locator('[hx-get*="questions"]').click();

  await expect(page.getByTestId('rating-questionnaire')).toBeVisible();
  await expect(page.getByTestId('rating-calculate')).toBeVisible();
});

test('Completing the questionnaire produces a score', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();
  await page.getByTestId('rating-confirm').locator('[hx-get*="questions"]').click();
  await expect(page.getByTestId('rating-questionnaire')).toBeVisible();

  // Answer every question by picking the first radio option in each fieldset
  const fieldsets = page.locator('[data-testid="rating-questionnaire"] fieldset');
  const count = await fieldsets.count();
  for (let i = 0; i < count; i++) {
    await fieldsets.nth(i).locator('input[type="radio"]').first().check();
  }

  await page.getByTestId('rating-calculate').click();

  await expect(page.getByTestId('rating-confirm')).toBeVisible();
  await expect(page.getByTestId('rating-input')).not.toHaveValue('');
});

test('Saving a rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();
  await expect(page.getByTestId('rating-confirm')).toBeVisible();

  await page.getByTestId('rating-input').fill('7');
  await page.getByTestId('rating-lock-in').click();

  await expect(page.locator('dialog[open]')).not.toBeVisible();
});

test('Deleting a rating', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-rating').locator('[hx-get*="rating-recommender"]').click();
  await expect(page.getByTestId('rating-confirm')).toBeVisible();

  await page.getByTestId('rating-delete').click();

  await expect(page.locator('dialog[open]')).not.toBeVisible();
});

test('Saving review notes', async ({ context, page }) => {
  expect(userId, 'E2E_TEST_USER_ID must be set').toBeTruthy();
  expect(albumId, 'E2E_TEST_ALBUM_ID must be set').toBeTruthy();

  await loginAs(context, userId!);
  await page.goto(`/app/library/albums/${albumId}`);

  await page.getByTestId('album-detail-notes').locator('button').click();
  await expect(page.locator('dialog[open]')).toBeVisible();

  await page.getByTestId('review-notes-textarea').fill('Great record.');
  await page.getByTestId('review-notes-save').click();

  await expect(page.locator('dialog[open]')).not.toBeVisible();
});
