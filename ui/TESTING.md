# UI Testing

Three layers. Prefer the lowest layer that can fail for the right reason.

| Layer           | Runner                   | Files                          | What belongs here                                               |
| --------------- | ------------------------ | ------------------------------ | --------------------------------------------------------------- |
| **Unit**        | Vitest `node`            | `src/**/*.test.ts`             | Pure libs, workers, composables with mocked Worker/window/fetch |
| **Integration** | Vitest `happy-dom` + Vue | `src/**/*.integration.test.ts` | Mount contracts: gates, retry, one settings write path          |
| **E2E**         | Playwright               | `e2e/**/*.spec.ts`             | User journeys against a served HTML fixture                     |

## Commands

```bash
pnpm test                 # unit + integration
pnpm test:unit            # unit only
pnpm test:integration     # integration only
pnpm test:e2e             # Playwright smokes (starts fixture server)
pnpm test:watch           # vitest watch (all vitest projects)
```

From repo root:

```bash
task test:ui              # unit + integration
task test:ui:e2e          # Playwright
```

## Shared fixtures

Import from `@/test-utils` (or relative `../test-utils`):

| Module                                 | Contents                                                                                            |
| -------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `dataFixtures`                         | `dp`, `noSort` / `ascSort` / `descSort`, `VALUE_AXES`, `MIXED_AXES`                                 |
| `datasetFixtures`                      | `ds(settings, data?)`                                                                               |
| `chartFixtures`                        | `baseConfig`, `make*ChartData`, `installDevicePixelRatio`, `groupedRender3D` / `continuousRender3D` |
| `workers/__test-utils__/workerHarness` | `installMockSelf`, `TrackedMockWorker`                                                              |

Do **not** copy-paste local `makeMixedConfig` / `dp` / DPR stubs into new suites.

## Rules

1. **New pure logic** → unit next to the module.
2. **New Vue interaction / gate** → one integration test.
3. **New user journey / prod regression** → one e2e smoke, not a suite explosion.
4. **Settings visibility** → `fieldRegistry.test.ts` only. Never re-assert `getRenderableFields` matrices elsewhere.
5. **Assert behavior** (axis type, stack on, URL key, button present) — not theme hex or font px.
6. **Pyramid budget** for a feature: ~10 units : 3 integrations : 1 e2e.
7. No faux component tests that mock every SFC then never mount.

## E2E fixtures

`e2e/fixtures/` holds static HTML that injects `window.VIZB_DATA`. Keep payloads tiny.
Stable selectors: prefer roles/labels; add `data-testid` only when needed.
