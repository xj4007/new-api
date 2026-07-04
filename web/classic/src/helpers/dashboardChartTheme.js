/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import VChart from '@visactor/vchart';
import { generateVChartSemiTheme } from '@visactor/vchart-semi-theme/cjs/generator';
import { observeAttribute } from '@visactor/vchart-theme-utils/cjs/utils/observe-attribute';

const THEME_MODE_ATTRIBUTE = 'theme-mode';
const LIGHT_THEME_NAME = 'semiDesignLight';
const DARK_THEME_NAME = 'semiDesignDark';

let initialized = false;
let watchingMode = false;

const getCurrentMode = () => {
  if (
    typeof document !== 'undefined' &&
    document.body.hasAttribute(THEME_MODE_ATTRIBUTE) &&
    document.body.getAttribute(THEME_MODE_ATTRIBUTE) === 'dark'
  ) {
    return 'dark';
  }

  return 'light';
};

const ensureThemeRegistered = (mode) => {
  const themeName = mode === 'dark' ? DARK_THEME_NAME : LIGHT_THEME_NAME;

  if (!VChart.ThemeManager.themeExist(themeName)) {
    VChart.ThemeManager.registerTheme(themeName, generateVChartSemiTheme(mode));
  }

  return themeName;
};

export const initClassicDashboardChartTheme = () => {
  if (typeof document === 'undefined' || typeof MutationObserver === 'undefined') {
    return;
  }

  const applyTheme = () => {
    const mode = getCurrentMode();
    const themeName = ensureThemeRegistered(mode);

    if (VChart.ThemeManager.getCurrentThemeName() !== themeName) {
      VChart.ThemeManager.setCurrentTheme(themeName);
    }
  };

  applyTheme();

  if (watchingMode) {
    initialized = true;
    return;
  }

  observeAttribute(document.body, THEME_MODE_ATTRIBUTE, () => {
    applyTheme();
  });

  watchingMode = true;
  initialized = true;
};

export const hasClassicDashboardChartThemeInitialized = () => initialized;
