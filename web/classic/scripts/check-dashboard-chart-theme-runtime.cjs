const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');

const helperPath = path.join(
  __dirname,
  '..',
  'src',
  'helpers',
  'dashboardChartTheme.js',
);
const hookPath = path.join(
  __dirname,
  '..',
  'src',
  'hooks',
  'dashboard',
  'useDashboardCharts.jsx',
);

const helperSource = fs.readFileSync(helperPath, 'utf8');
const hookSource = fs.readFileSync(hookPath, 'utf8');

assert(
  helperSource.includes("import VChart from '@visactor/vchart';"),
  'dashboardChartTheme must bind to classic @visactor/vchart directly.',
);

assert(
  helperSource.includes("import { generateVChartSemiTheme } from '@visactor/vchart-semi-theme/cjs/generator';"),
  'dashboardChartTheme must reuse the semi theme generator without importing the theme package runtime entry.',
);

assert(
  !hookSource.includes("from '@visactor/vchart-semi-theme'"),
  'useDashboardCharts must not import the vchart-semi-theme runtime entry directly.',
);

const requireFromHook = createRequire(hookPath);
const vchartEntry = requireFromHook.resolve('@visactor/vchart');
const generatorEntry = requireFromHook.resolve('@visactor/vchart-semi-theme/cjs/generator');
const generatorRequire = createRequire(generatorEntry);

assert.equal(
  generatorRequire.resolve('@visactor/vchart'),
  path.join(
    path.dirname(requireFromHook.resolve('@visactor/vchart-semi-theme')),
    '..',
    'node_modules',
    '@visactor',
    'vchart',
    'cjs',
    'index.js',
  ),
  'The semi theme package still resolves its own nested VChart runtime, which is why classic must not call its runtime entry.',
);

assert.equal(
  requireFromHook.resolve('@visactor/vchart'),
  vchartEntry,
  'Classic hook should resolve a stable direct VChart runtime.',
);

console.log('dashboard chart theme runtime check passed');
