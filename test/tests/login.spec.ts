import { expect, test } from '@playwright/test';

test('log in as user', async ({ page }) => {
  await page.goto('http://localhost:8080/');
  await expect(page).toHaveURL('http://localhost:8080/login');
  await page.getByLabel('Nutzername').fill('fry');
  await page.getByLabel('Nutzername').press('Tab');
  await page.getByLabel('Passwort').fill('fry');
  await page.getByLabel('Passwort').press('Enter');
  await expect(page).toHaveURL('http://localhost:8080/aussonderungen');
  await expect(page.getByRole('button', { name: 'Philip J. Fry' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Abmelden' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Steuerungsstelle' })).toHaveCount(0);
  await expect(page.getByRole('link', { name: 'Administration' })).toHaveCount(0);
  await expect(page.getByLabel('Für alle Mitarbeiter anzeigen')).toHaveCount(0);
});

test('log in as admin', async ({ page }) => {
  await page.goto('http://localhost:8080/');
  await expect(page).toHaveURL('http://localhost:8080/login');
  await page.getByLabel('Nutzername').fill('hermes');
  await page.getByLabel('Nutzername').press('Tab');
  await page.getByLabel('Passwort').fill('hermes');
  await page.getByLabel('Passwort').press('Enter');
  await expect(page).toHaveURL('http://localhost:8080/aussonderungen');
  await expect(page.getByRole('button', { name: 'Hermes Conrad' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Abmelden' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Steuerungsstelle' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Administration' })).toBeVisible();
  await expect(page.getByLabel('Für alle Mitarbeiter anzeigen')).toBeVisible();
});
