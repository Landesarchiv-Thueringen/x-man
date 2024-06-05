import { expect, test } from '@playwright/test';
import { uploadFile } from './webdav';

// The Setup step has to be performed only once, but can safely be run multiple
// times.
test.describe('setup', () => {
  test('log in as Hermes', async ({ page }) => {
    await page.goto('http://localhost:8080/');
    await page.getByLabel('Nutzername').fill('hermes');
    await page.getByLabel('Nutzername').press('Tab');
    await page.getByLabel('Passwort').fill('hermes');
    await page.getByLabel('Passwort').press('Enter');
    await expect(page).toHaveURL('http://localhost:8080/aussonderungen');
    await page.context().storageState({ path: 'playwright/auth/hermes.json' });
  });

  test.describe('Hermes', () => {
    test.use({ storageState: 'playwright/auth/hermes.json' });

    test('ensure Fry is archivist for TSK', async ({ page }) => {
      await page.goto('http://localhost:8080/administration/abgebende-stellen');
      await expect(page.getByRole('cell', { name: 'Thüringer Staatskanzlei' })).toBeVisible();
      // Check current values of TSK
      const texts = await page.getByRole('row', { name: 'Thüringer Staatskanzlei' }).innerText();
      if (!texts.includes('Philip J. Fry')) {
        // Register Fry as archivist for TSK
        await page.getByRole('cell', { name: 'Thüringer Staatskanzlei' }).allInnerTexts();
        await page.getByRole('cell', { name: 'Thüringer Staatskanzlei' }).getByLabel('Details anzeigen').click();
        await page.getByPlaceholder('Filtern').click();
        await page.getByRole('option', { name: 'Philip J. Fry' }).click();
        await page.getByRole('button', { name: 'Speichern' }).click();
      }
      await expect(page.getByRole('row', { name: 'Thüringer Staatskanzlei' })).toContainText('Philip J. Fry');
    });
  });

  test('log in as Fry', async ({ page }) => {
    await page.goto('http://localhost:8080/');
    await page.getByLabel('Nutzername').fill('fry');
    await page.getByLabel('Nutzername').press('Tab');
    await page.getByLabel('Passwort').fill('fry');
    await page.getByLabel('Passwort').press('Enter');
    await expect(page).toHaveURL('http://localhost:8080/aussonderungen');
    await page.context().storageState({ path: 'playwright/auth/fry.json' });
  });
});

// Note that `page` has to be passed even though we don't use it.
test('upload 0501 message', async ({ page }) => {
  await uploadFile('test-data/9a75050f-323a-4e84-94c9-a889aa2b4fe8_Aussonderung.Anbieteverzeichnis.0501.zip');
});

test.describe('Fry', () => {
  test.use({ storageState: 'playwright/auth/fry.json' });

  test('find message in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row).toContainText('Thüringer Staatskanzlei');
    await expect(row.getByText('check')).toHaveCount(1);
    await expect(page.getByLabel('Details anzeigen')).toHaveAttribute(
      'href',
      '/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8'
    );
  });

  test('appraise message', async ({ page }) => {
    await page.goto('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8/0501');
    await page.getByRole('button', { name: 'Mehrfachauswahl' }).click();
    await page.getByLabel('Akte: 1234-1').check();
    await page.getByLabel('Akte: 1234-2').check();
    await page.getByRole('button', { name: 'Bewerten' }).click();
    await page.getByText('Bewertungsentscheidung', { exact: true }).click();
    await page.getByRole('option', { name: 'Archivieren' }).click();
    await page.getByRole('button', { name: 'Speichern' }).click();
    await expect(page.getByText('Bewertung erfolgreich gespeichert')).toBeVisible();
    await expect(page.locator('mat-panel-description')).toContainText('Bewertung (2 / 3)');
  });

  test('partial appraisal in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row.getByText('check')).toHaveCount(1);
    await expect(row.locator('mat-cell.mat-column-appraisalComplete')).toContainText('2 / 3');
  });

  test('send appraisal', async ({ page }) => {
    await page.goto('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8/0501');
    await page.getByRole('button', { name: 'Bewertung senden' }).click();
    await page.getByRole('button', { name: 'Bewertung senden' }).click();
    await expect(page.getByText('Bewertungsnachricht wurde erfolgreich versandt')).toBeVisible();
    await expect(page.locator('mat-panel-description')).toContainText('Bewertung abgeschlossen');
  });

  test('complete appraisal in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row.getByText('check')).toHaveCount(2);
  });
});

test('upload 0505 message', async ({ page }) => {
  await uploadFile('test-data/9a75050f-323a-4e84-94c9-a889aa2b4fe8_Aussonderung.BewertungEmpfangBestaetigen.0505.zip');
});

test.describe('Fry', () => {
  test.use({ storageState: 'playwright/auth/fry.json' });

  test('appraisal ack in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row.getByText('check')).toHaveCount(3);
  });
});

test('upload 0503 message', async ({ page }) => {
  await uploadFile('test-data/9a75050f-323a-4e84-94c9-a889aa2b4fe8_Aussonderung.Aussonderung.0503.zip');
});

test.describe('Fry', () => {
  test.use({ storageState: 'playwright/auth/fry.json' });

  test('received 0503 in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row.getByText('check')).toHaveCount(4);
  });

  test('archive message', async ({ page }) => {
    await page.goto('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    await expect(page).toHaveURL('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8/0503/details');
    await page.getByRole('button', { name: 'Abgabe archivieren' }).click();
    await page.getByRole('button', { name: 'Archivierungsprozess starten' }).click();
    await expect(page.getByText('Archivierung gestartet')).toBeVisible();
    await expect(page.locator('mat-panel-description')).toContainText('Abgabe archiviert');
  });

  test('archived in table', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.getByRole('button', { name: 'Filtern' }).click();
    await page.getByLabel('Filtern').fill('9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const row = page.locator('mat-row');
    await expect(row.getByText('check')).toHaveCount(5);
  });

  test('download report', async ({ page }) => {
    await page.goto('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Übernahmebericht herunterladen' }).click();
    await downloadPromise;
  });
});

test.describe('Hermes', () => {
  test.use({ storageState: 'playwright/auth/hermes.json' });

  test.skip('delete process', async ({ page }) => {
    await page.goto('http://localhost:8080/nachricht/9a75050f-323a-4e84-94c9-a889aa2b4fe8');
    await page.getByRole('button', { name: 'Administration' }).click();
    await page.getByRole('button', { name: 'Aussonderung löschen' }).click();
    await page.getByRole('button', { name: 'Löschen' }).click();
    await expect(page.getByText('Aussonderung gelöscht')).toBeVisible();
  });
});
