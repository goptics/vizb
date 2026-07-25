import { test as it, expect } from '@playwright/test'

it.describe('deep link', () => {
  it('applies ?c=line as the active chart type', async ({ page }) => {
    await page.goto('/?c=line')

    const active = page.getByTestId('chart-type-active')
    await expect(active).toBeVisible()
    await expect(active).toHaveAttribute('data-chart-type', 'line')
    await expect(active).toContainText('line')
    await expect(page.getByTestId('chart-plot')).toContainText('line chart')
  })
})
