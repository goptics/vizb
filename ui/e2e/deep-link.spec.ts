import { test, expect } from '@playwright/test'

test.describe('deep link', () => {
  test('applies ?c=line as the active chart type', async ({ page }) => {
    await page.goto('/?c=line')

    const active = page.getByTestId('chart-type-active')
    await expect(active).toBeVisible()
    await expect(active).toHaveAttribute('data-chart-type', 'line')
    await expect(active).toContainText('line')
    await expect(page.getByTestId('chart-plot')).toContainText('line chart')
  })
})
