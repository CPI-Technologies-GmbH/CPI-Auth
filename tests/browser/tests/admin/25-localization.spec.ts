import { test, expect } from '@playwright/test';
import { goToAdmin, screenshot, waitForLoad, assertNoErrors } from './helpers';

test.describe('Localization / Language Switcher', () => {
  test('should show language switcher in header', async ({ page }) => {
    await goToAdmin(page, '/');
    await screenshot(page, '25-dashboard-default-lang');
    await assertNoErrors(page);

    // Language switcher button should be visible (shows EN/DE/FR/ES)
    const langBtn = page.locator('header button').filter({ hasText: /^EN$|^DE$|^FR$|^ES$/ });
    await expect(langBtn).toBeVisible({ timeout: 5000 });
  });

  test('should switch language to German', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);

    // Click language switcher
    const langBtn = page.locator('header button').filter({ hasText: /^EN$|^DE$|^FR$|^ES$/ });
    await langBtn.click();
    await page.waitForTimeout(300);

    // Select Deutsch from the dropdown
    const deutschOption = page.getByRole('button', { name: /Deutsch/ });
    await expect(deutschOption).toBeVisible({ timeout: 3000 });
    await deutschOption.click();
    await page.waitForTimeout(500);
    await screenshot(page, '25-dashboard-german');
    await assertNoErrors(page);

    // Navigate to users page - should show German nav
    await page.goto('/users', { waitUntil: 'domcontentloaded' });
    await waitForLoad(page);
    await screenshot(page, '25-users-german');

    // Reset to English
    const langBtn2 = page.locator('header button').filter({ hasText: /^EN$|^DE$|^FR$|^ES$/ });
    await langBtn2.click();
    await page.waitForTimeout(300);
    await page.getByRole('button', { name: /English/ }).click();
    await page.waitForTimeout(300);
  });

  test('should persist language selection across navigation', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);

    // Switch to French
    const langBtn = page.locator('header button').filter({ hasText: /^EN$|^DE$|^FR$|^ES$/ });
    await langBtn.click();
    await page.waitForTimeout(300);
    await page.getByRole('button', { name: /Francais/ }).click();
    await page.waitForTimeout(500);

    // Navigate to settings
    await page.goto('/settings', { waitUntil: 'domcontentloaded' });
    await waitForLoad(page);
    await screenshot(page, '25-settings-french');
    await assertNoErrors(page);

    // Language should still be French (FR shown in switcher)
    const switcherText = page.locator('header button').filter({ hasText: /^FR$/ });
    await expect(switcherText).toBeVisible({ timeout: 3000 });

    // Reset to English
    await switcherText.click();
    await page.waitForTimeout(300);
    await page.getByRole('button', { name: /English/ }).click();
    await page.waitForTimeout(300);
  });

  test('should show all 4 language options', async ({ page }) => {
    await goToAdmin(page, '/');
    await waitForLoad(page);

    // Open language dropdown
    const langBtn = page.locator('header button').filter({ hasText: /^EN$|^DE$|^FR$|^ES$/ });
    await langBtn.click();
    await page.waitForTimeout(300);
    await screenshot(page, '25-language-dropdown-open');

    // All 4 languages should be listed as buttons in the dropdown
    await expect(page.getByRole('button', { name: /English/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Deutsch/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Francais/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Espanol/ })).toBeVisible();
  });
});
