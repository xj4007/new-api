# Classic Dashboard VChart Runtime Fix

## Purpose

This note records the root cause and code-level fix for the `classic` frontend dashboard render failure:

- Symptom: opening `数据看板` in `经典前端` crashed the page
- Browser error: `TypeError: Cannot read properties of undefined (reading 'createCanvas')`

Use this file as merge context when:

- merging `main`
- doing secondary development on `web/classic`
- upgrading VisActor chart dependencies
- adjusting classic build/install flow

## What Broke

The failure was not caused by the Boxmoji theme worktree. After reverting to the original classic source, the dashboard still crashed.

The actual issue was a split chart runtime in the classic dependency graph:

1. `web/classic/src/hooks/dashboard/useDashboardCharts.jsx` originally imported `initVChartSemiTheme` from `@visactor/vchart-semi-theme`.
2. The actual chart rendering path used `@visactor/react-vchart` from the classic workspace.
3. `@visactor/vchart-semi-theme` resolved through `web/node_modules/...` and also pulled its own nested `@visactor/vchart`.
4. `@visactor/react-vchart` and `@visactor/vchart` also resolved their own nested `@visactor/vrender-core` and `@visactor/vrender-kits`.
5. As a result, theme registration and browser canvas environment setup did not always happen on the same singleton graph that the rendered charts used.

That mismatch is what led to `createCanvas` failing at runtime.

## Why Original Code Still Failed

The classic dashboard source matched `origin/main`, but the installed dependency tree did not behave like a single-runtime setup.

So this was a case of:

- source code looks original
- runtime graph is not effectively original

That is why "reverting code" did not remove the error.

## Why Another Remote/Mirror Build Might Look Fine

This issue is sensitive to dependency installation layout and build flow.

The repository already contains evidence that classic installs are special:

- commit `c12e5db4` changed release/CI behavior to install classic workspace dependencies separately
- local release build also installs classic separately after default

That means two environments can use the same source but end up with different dependency resolution layouts.

If another remote image or mirror build did not reproduce the bug, the most likely reason is:

- its install graph did not split the VisActor runtime in the same way

## Files Changed For The Fix

### `web/classic/src/hooks/dashboard/useDashboardCharts.jsx`

- Removed the direct `@visactor/vchart-semi-theme` runtime entry import
- Replaced `initVChartSemiTheme(...)` with a classic-local initializer:
  `initClassicDashboardChartTheme()`

Why this matters:

- this file is the original root of the dashboard chart theme initialization
- if a future merge restores direct `initVChartSemiTheme(...)`, the crash can return

### `web/classic/src/helpers/dashboardChartTheme.js`

New helper added to keep theme registration on the same `@visactor/vchart` runtime used by classic charts.

It does three things:

1. imports classic `@visactor/vchart`
2. imports only the semi theme generator data, not the theme package runtime entry
3. watches `body[theme-mode]` and updates `VChart.ThemeManager` on the classic runtime

Why this matters:

- this file is the main bugfix
- if future code moves chart theme logic elsewhere, preserve the "same runtime" rule

### `web/classic/rsbuild.config.ts`

Added aliases for:

- `@visactor/vrender-core`
- `@visactor/vrender-kits`

These aliases point to the copies under classic `@visactor/vchart`'s own dependency tree.

Why this matters:

- even after fixing theme initialization, the classic chart stack could still split at the lower `vrender-*` level
- these aliases force the rendered chart path to share the same runtime singletons

### `web/classic/scripts/check-dashboard-chart-theme-runtime.cjs`

New diagnostic guard script.

Purpose:

- verifies the classic dashboard no longer imports `@visactor/vchart-semi-theme` runtime entry directly
- documents that the semi theme package still resolves its own nested chart runtime
- helps quickly re-check this bug after merges or dependency upgrades

## Merge / Secondary Development Guidance

If you later merge `main` and hit conflicts, pay special attention to these files:

- `web/classic/src/hooks/dashboard/useDashboardCharts.jsx`
- `web/classic/src/helpers/dashboardChartTheme.js`
- `web/classic/rsbuild.config.ts`
- `web/classic/scripts/check-dashboard-chart-theme-runtime.cjs`

### Conflict Resolution Rule

Do not blindly restore:

```js
import { initVChartSemiTheme } from '@visactor/vchart-semi-theme';
```

and do not remove the `vrender` aliases unless you have verified that the classic chart runtime is truly unified again.

### If `main` Changes Dashboard Chart Logic

Keep the upstream data/spec logic, but preserve these two behaviors:

1. classic dashboard theme registration must stay on the same `@visactor/vchart` runtime used by chart rendering
2. classic build must keep `@visactor/vrender-core` and `@visactor/vrender-kits` unified through aliasing, unless dependency layout is proven safe without it

### If `main` Upgrades VisActor Packages

Re-check all of the following before deleting this fix:

1. whether `@visactor/vchart-semi-theme` still resolves its own nested `@visactor/vchart`
2. whether `@visactor/react-vchart` still resolves nested `@visactor/vrender-core` / `@visactor/vrender-kits`
3. whether `数据看板` opens without `createCanvas` errors after a fresh build and deploy

Only remove this fix when the runtime graph is confirmed unified, not just because versions "look aligned".

## Verification Used For This Fix

Local verification:

- `node web/classic/scripts/check-dashboard-chart-theme-runtime.cjs`
- `bun run build` in `web/classic`
- `powershell -File deploy/production/app/build-local-release.ps1`

Deployment verification:

- `python deploy/production/app/publish-local-release.py`
- container health check passed
- `/api/status` returned success
- online `/console` bundle hash changed after deployment

Functional result:

- after the final runtime-unification fix, `经典前端 -> 数据看板` rendered normally again

## Short Summary

This was a runtime-singleton split bug, not a dashboard business-logic bug.

The durable lesson is:

- classic chart theme setup, chart rendering, and low-level `vrender-*` runtime must stay on one shared VisActor runtime chain

If that invariant is broken in a future merge, the `createCanvas` crash is likely to come back.
