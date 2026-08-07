---
kind: frontend_style
name: Tailwind v4 + shadcn/ui Design System with CSS Token Theming
category: frontend_style
scope:
    - '**'
source_files:
    - frontend/src/app/globals.css
    - frontend/components.json
    - frontend/postcss.config.mjs
    - frontend/package.json
    - frontend/src/lib/utils.ts
    - frontend/src/components/ui/button.tsx
    - frontend/src/components/ui/badge.tsx
    - frontend/src/app/layout.tsx
---

The frontend uses a modern, token-driven styling stack built on Tailwind CSS v4 and the shadcn/ui component library, layered with a production-grade CSS custom-property design system.

Core stack
- Tailwind CSS v4 via @tailwindcss/postcss (no legacy tailwind.config.js; styles declared in src/app/globals.css).
- shadcn/ui configured for Next.js App Router (components.json: style new-york, RSC enabled, aliases @/components/ui, icon library lucide-react).
- Class composition via clsx + tailwind-merge through a shared cn() helper in src/lib/utils.ts.
- Variant-driven components using class-variance-authority (CVA) — see src/components/ui/button.tsx, badge.tsx.
- Radix UI primitives (@radix-ui/react-slot, @radix-ui/react-dropdown-menu) for unstyled, accessible building blocks.
- Motion animations via motion (Framer Motion).

Design tokens & theming
All visual tokens live as CSS custom properties in :root and :root[data-theme="dark"] inside globals.css:
- Color palette: --primary, --secondary, --accent, --destructive, --muted-bg, --border, --input, --ring, plus semantic surface tokens (--panel, --panel-soft, --panel-strong, --bg, --text, --text-soft, --line).
- Typography scale: fluid clamp()-based tokens --text-xs ... --text-5xl, font families --font-sans, --font-display, --font-thai (Inter, Outfit, Noto Sans Thai loaded via next/font).
- Radius scale (--radius-xs ... --radius-full), shadow scale (--shadow-xs ... --shadow-xl, --shadow-glow), motion tokens (--ease-*, --duration-*).
- Density tokens (--density-page-pad, --density-card-pad, --density-control-h, --density-gap, ...) used throughout login/workspace layouts.
- Login-shell-specific tokens (--login-shell-bg, --login-glass-bg*, --login-form-shadow*, --login-primary-glow*) drive the glassmorphic two-panel login screen.
- Theme mode is persisted to cookies and applied via data-theme / data-color-theme attributes on <html>; layout.tsx merges cookie-backed theme variables into inline style props so there is no FOUC.

CSS methodology
- Global base reset and utility classes in globals.css override Tailwind defaults where needed (e.g. .truncate forced visible per project rule, input backgrounds set transparent).
- A single @custom-variant dark (&:where([data-theme="dark"], [data-theme="dark"] *)) enables dark-mode variants without relying on prefers-color-scheme alone.
- The @theme inline { ... } block maps shadcn semantic color names to the CSS variables above, keeping component classes decoupled from concrete colors.
- Responsive behavior is achieved with Tailwind's responsive prefixes and clamp()-based fluid typography/density rather than fixed breakpoints.

Component conventions
- All reusable UI lives under src/components/ui/* and follows the shadcn pattern: one file per primitive, CVA variants for variant/size, forwardRef, typed props extending VariantProps<typeof XVariants>, and cn(...) for class merging.
- Business-facing screens compose these primitives alongside domain-specific components under src/components/product-barcode/....
- Icons come from lucide-react; charts from recharts; maps from leaflet + react-leaflet.

What developers should follow
- Use the cn() helper for all dynamic className composition; never concatenate strings manually.
- Prefer shadcn semantic tokens (bg-primary, text-muted-foreground, border-border, shadow-card) over raw hex values.
- When adding new colors or sizes, extend the CSS variable sets in globals.css (:root and :root[data-theme="dark"]) and map them via @theme inline if they need to be consumed by Tailwind utilities.
- Keep variant definitions in CVA within each src/components/ui/*.tsx file; do not scatter ad-hoc style rules in page components.
- For layout density, reach for --density-* tokens instead of hard-coded pixel paddings.