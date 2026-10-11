// Implemented menu screens (not "รอพัฒนา"), read from the real frontend catalog so the UI smoke never drifts
// from the menu: MENU_BARS → menuBarEntries() → drop isMenuScreenPending(route).
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const FRONTEND = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../frontend');

export function implementedScreens() {
  const require = createRequire(path.join(FRONTEND, 'package.json'));
  const { createJiti } = require('jiti');
  const jiti = createJiti(path.join(FRONTEND, 'index.js'), { fsCache: false, alias: { '@': path.join(FRONTEND, 'src') } });
  const { MENU_BARS, menuBarEntries } = jiti(path.join(FRONTEND, 'src/lib/menu-bars.ts'));
  const { isMenuScreenPending, isMenuDataPending } = jiti(path.join(FRONTEND, 'src/lib/menu-screen-status.ts'));
  const { MENU_SECTIONS } = jiti(path.join(FRONTEND, 'src/lib/menu-data.ts'));
  // The menu search lists MENU_SECTIONS only; bar-only screens are opened from their menu-bar head instead.
  const searchable = new Set(MENU_SECTIONS.flatMap((s) => s.groups.flatMap((g) => g.items.map((i) => i.route))));
  const seen = new Set();
  const screens = [];
  for (const bar of MENU_BARS) {
    for (const item of menuBarEntries(bar.nodes)) {
      if (!item.route || seen.has(item.route) || isMenuScreenPending(item.route)) continue;
      seen.add(item.route);
      screens.push({ head: bar.id, id: item.id, route: item.route, label: item.label?.th ?? item.id, dataPending: isMenuDataPending(item.route), searchable: searchable.has(item.route) });
    }
  }
  return screens;
}
