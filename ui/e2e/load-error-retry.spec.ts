import { test as it, expect } from '@playwright/test'

it.describe('load error retry', () => {
  it('Retry clears error and shows charts', async ({ page }) => {
    await page.goto('/?error=1')

    await expect(page.getByTestId('load-error')).toBeVisible()
    await expect(page.getByTestId('dashboard')).toBeHidden()

    await page.getByTestId('retry-button').click()

    await expect(page.getByTestId('load-error')).toBeHidden()
    await expect(page.getByTestId('chart-card')).toBeVisible()
    await expect(page.getByTestId('dataset-name')).toHaveText('Alpha Benchmark')
    await expect(page).toHaveURL(/^(?!.*error=1).*$/)
  })
})
