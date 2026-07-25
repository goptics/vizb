import { test as it, expect } from '@playwright/test'

it.describe('stats csv', () => {
  it('opening stats and clicking CSV fires a download', async ({ page }) => {
    await page.goto('/')

    await page.getByTestId('stats-button').click()
    await expect(page.getByTestId('stats-panel')).toBeVisible()

    const downloadPromise = page.waitForEvent('download')
    await page.getByTestId('csv-download').click()
    const download = await downloadPromise

    expect(download.suggestedFilename()).toBe('revenue-descriptive.csv')
  })
})
