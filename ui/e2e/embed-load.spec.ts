import { test, expect } from '@playwright/test'

test.describe('embed load', () => {
  test('shows dataset name and chart card from VIZB_DATA', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByTestId('dataset-name')).toHaveText('Alpha Benchmark')
    await expect(page.getByTestId('chart-card')).toBeVisible()
    await expect(page.getByTestId('chart-title')).toHaveText('revenue')
    await expect(page.getByTestId('load-error')).toBeHidden()
  })
})
