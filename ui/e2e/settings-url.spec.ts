import { test, expect } from '@playwright/test'

test.describe('settings url', () => {
  test('toggling stack writes expected URL param', async ({ page }) => {
    await page.goto('/')

    await page.getByTestId('settings-toggle').click()
    await expect(page.getByTestId('settings-panel')).toBeVisible()

    await page.getByTestId('stack-toggle').click()

    await expect(page).toHaveURL(/\bbar\.st=true\b/)
    await expect(page.getByTestId('url-display')).toContainText('bar.st=true')
    await expect(page.getByTestId('stack-toggle')).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByTestId('chart-plot')).toContainText('stacked')
  })
})
