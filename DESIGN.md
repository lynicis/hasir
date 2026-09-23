---
version: alpha
name: HASIR Design System
colors:
  # Dashboard UI Colors
  background: "oklch(0.99 0.003 240)"
  foreground: "oklch(0.12 0.01 240)"
  card: "oklch(1 0 0)"
  card-foreground: "oklch(0.12 0.01 240)"
  popover: "oklch(1 0 0)"
  popover-foreground: "oklch(0.12 0.01 240)"
  primary: "oklch(0.25 0.02 240)"
  primary-foreground: "oklch(0.99 0.003 240)"
  secondary: "oklch(0.95 0.005 240)"
  secondary-foreground: "oklch(0.25 0.02 240)"
  muted: "oklch(0.95 0.005 240)"
  muted-foreground: "oklch(0.45 0.015 240)"
  accent: "oklch(0.95 0.005 240)"
  accent-foreground: "oklch(0.25 0.02 240)"
  destructive: "oklch(0.55 0.22 25)"
  border: "oklch(0.9 0.005 240)"
  input: "oklch(0.9 0.005 240)"
  ring: "oklch(0.6 0.02 240)"
  
  # Dark mode variants
  dark-background: "oklch(0.08 0.005 240)"
  dark-foreground: "oklch(0.98 0.005 240)"
  dark-card: "oklch(0.12 0.008 240)"
  dark-card-foreground: "oklch(0.98 0.005 240)"
  dark-primary: "oklch(0.92 0.005 240)"
  dark-primary-foreground: "oklch(0.12 0.008 240)"
  
  # Sidebar tokens
  sidebar: "oklch(0.99 0.003 240)"
  sidebar-foreground: "oklch(0.12 0.01 240)"
  sidebar-primary: "oklch(0.25 0.02 240)"
  sidebar-primary-foreground: "oklch(0.99 0.003 240)"
  sidebar-accent: "oklch(0.95 0.005 240)"
  sidebar-accent-foreground: "oklch(0.25 0.02 240)"
  sidebar-border: "oklch(0.9 0.005 240)"
  sidebar-ring: "oklch(0.6 0.02 240)"

  # Chart tokens
  chart-1: "oklch(0.646 0.222 41.116)"
  chart-2: "oklch(0.6 0.118 184.704)"
  chart-3: "oklch(0.398 0.07 227.392)"
  chart-4: "oklch(0.828 0.189 84.429)"
  chart-5: "oklch(0.769 0.188 70.08)"

typography:
  # Dashboard Typography
  hero:
    fontFamily: Geist Sans
    fontSize: 6rem
    fontWeight: 800
    letterSpacing: -0.05em
  page-title:
    fontFamily: Geist Sans
    fontSize: 1.25rem
    fontWeight: 600
  card-title:
    fontFamily: Geist Sans
    fontWeight: 500
  dialog-title:
    fontFamily: Geist Sans
    fontSize: 1.125rem
    fontWeight: 600
  body:
    fontFamily: Geist Sans
    fontSize: 0.875rem
    fontWeight: 400
  button:
    fontFamily: Geist Sans
    fontSize: 0.875rem
    fontWeight: 500
  label:
    fontFamily: Geist Sans
    fontSize: 0.875rem
    fontWeight: 500
  code:
    fontFamily: Geist Mono
    fontSize: 0.75rem
  error:
    fontFamily: Geist Sans
    fontSize: 0.875rem
    fontWeight: 400

rounded:
  # Dashboard Rounding
  sm: 8px
  md: 10px
  lg: 12px
  xl: 16px
  card: 4px      # --radius-card: sharp 4px for all cards
  button: 0px    # --radius-button: fully sharp buttons

spacing:
  # Dashboard Spacing
  xs: 4px
  sm: 8px
  md: 16px
  lg: 32px
  xl: 64px
  gap-card: 16px   # card-spacing (--spacing(4))
  padding-card: 16px

components:
  button-default:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.button}"
    height: 32px
  button-destructive:
    backgroundColor: "{colors.destructive}/10"
    textColor: "{colors.destructive}"
    rounded: "{rounded.button}"
    height: 32px
  button-outline:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.button}"
    height: 32px
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-foreground}"
    rounded: "{rounded.button}"
    height: 32px
  button-ghost:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.button}"
    height: 32px
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    rounded: "{rounded.card}"
    ring: "ring-1 ring-foreground/10"
  dialog:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    rounded: "{rounded.lg}"
    padding: 24px
  popover:
    backgroundColor: "{colors.popover}"
    textColor: "{colors.popover-foreground}"
    rounded: "{rounded.md}"
  input:
    backgroundColor: "transparent"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    height: 32px
  border-indicator:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.border}"
  badge-muted:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.muted-foreground}"
  sidebar:
    backgroundColor: "{colors.sidebar}"
    textColor: "{colors.sidebar-foreground}"
  sidebar-primary:
    backgroundColor: "{colors.sidebar-primary}"
    textColor: "{colors.sidebar-primary-foreground}"
  sidebar-item-active:
    backgroundColor: "{colors.sidebar-accent}"
    textColor: "{colors.sidebar-accent-foreground}"
  focus-ring:
    backgroundColor: "{colors.ring}"
  sidebar-focus-ring:
    backgroundColor: "{colors.sidebar-ring}"
  sidebar-border:
    backgroundColor: "{colors.sidebar-border}"
  chart-legend-1:
    backgroundColor: "{colors.chart-1}"
  chart-legend-2:
    backgroundColor: "{colors.chart-2}"
  chart-legend-3:
    backgroundColor: "{colors.chart-3}"
  chart-legend-4:
    backgroundColor: "{colors.chart-4}"
  chart-legend-5:
    backgroundColor: "{colors.chart-5}"
  card-dark:
    backgroundColor: "{colors.dark-card}"
    textColor: "{colors.dark-card-foreground}"
  button-default-dark:
    backgroundColor: "{colors.dark-primary}"
    textColor: "{colors.dark-primary-foreground}"
  page-dark:
    backgroundColor: "{colors.dark-background}"
    textColor: "{colors.dark-foreground}"
---

# Design System: HASIR
**Project ID:** c5027631-e040-4ab2-b704-c4ac4ebef772

## 1. Visual Theme & Atmosphere
Hasir is a self-hosted protobuf schema registry built for engineers. The design system reflects a commit-as-currency aesthetic: terminal-forward, monospace-led brutalism. It aims to project enterprise trust through precision and restraint, eschewing marketing fluff for signal clarity. The visual language is anchored to a near-black void with a subtle warm blue undertone ("dark stage"), contrasted by pure-ish white typography and a singular warm amber signal ("bright answer") that echoes the legacy branding as a terminal cursor or commit diff line.

Hasir Dashboard provides a clean, premium, and highly functional management interface for the Hasir schema registry.

### Technology Stack

| Layer | Technology | Version |
|---|---|---|
| Framework | Next.js (App Router) | 16.3.5 |
| UI Library | React | 19.3.0 |
| Language | TypeScript | 6.x (strict) |
| Package Manager | Bun | ~1.4 |
| Styling | Tailwind CSS | 4.x |
| Component System | shadcn/ui (new-york style) | 4.x |
| Primitives | **@base-ui/react** | 1.8.x |
| Variant Management | class-variance-authority (CVA) | 0.7.x |
| Class Merging | clsx + tailwind-merge via `cn()` | — |
| Icons | Lucide React | 1.27.x |
| Animation | tw-animate-css + Framer Motion | 12.x |
| Forms | react-hook-form + @hookform/resolvers | 7.x / 5.x |
| Validation | Zod v4 | 4.x |
| Data Fetching | ConnectRPC + @tanstack/react-query | 2.x / 5.x |
| Client State | Zustand | 5.x |
| Theming | next-themes | 0.4.x |
| Toasts | Sonner | 2.x |
| Dates | Luxon | 3.x |
| Markdown | react-markdown + remark-gfm + rehype-raw/sanitize | — |
| Syntax Highlighting | react-syntax-highlighter (vscDarkPlus) | 16.x |

> [!IMPORTANT]
> **@base-ui/react** is the primitive layer for `Button` and `Input`. It replaces the previous direct Radix UI usage for those components. All other complex primitives (`DropdownMenu`, `Dialog`, `Tabs`, etc.) continue to use Radix UI via the `radix-ui` umbrella package.

### Configuration

**shadcn/ui** (`components.json`):

```json
{
  "style": "new-york",
  "rsc": true,
  "tsx": true,
  "tailwind": {
    "css": "../../packages/ui/src/styles/dashboard.css",
    "baseColor": "neutral",
    "cssVariables": true
  },
  "iconLibrary": "lucide",
  "aliases": {
    "components": "@/components",
    "utils": "@hasir/ui/lib/utils",
    "ui": "@hasir/ui/components",
    "lib": "@/lib",
    "hooks": "@/hooks"
  },
  "registries": {
    "@magicui": "https://magicui.design/r/{name}.json"
  }
}
```

> [!NOTE]
> The global CSS entry point has moved to `packages/ui/src/styles/dashboard.css` (exported as `@hasir/ui/dashboard.css`). The dashboard's `app/layout.tsx` imports it from the shared UI package — not from a local `app/globals.css`.

### Utility Function

All components use a single class-merge utility:

```typescript
// lib/utils.ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

## 2. Color Palette & Roles
* **Void Canvas (`#0a0b0d`)** - Canvas (`--bg`): Warm-undertone near-black. Default canvas for all new UI.
* **Surface Panel (`#101213`)** - Surface (`--surface`): Slightly elevated panels for cards and containers.
* **White Signal (`#f8f8f8`)** - Foreground (`--fg`): Pure-ish white for high-signal readable text and labels.
* **Muted Text (`#8f8f8f`)** - Muted (`--muted`): Ash grey for secondary interface text, helper copy, and metadata.
* **Hairline Border (`#202223`)** - Border (`--border`): Hairline structure for dividers and outlines.
* **Warm Amber Accent (`#eba941`)** - Signal Accent (`--accent`): Warm amber cursor accent. Used sparingly (≤2 times per screen) for critical statuses or key visual anchors.
* **Primary Action Background (`#ffffff`)** - Primary Action (`--accent-secondary`): Pure white. Used for primary CTAs requiring dark text on filled backgrounds for maximum contrast.

All color tokens use the **OKLCH** color space. The system uses CSS custom properties consumed by Tailwind via `@theme inline`.

### Light Mode (`:root`)

- **Background (`oklch(0.99 0.003 240)`):** Page background.
- **Foreground (`oklch(0.12 0.01 240)`):** Primary text.
- **Card (`oklch(1 0 0)`):** Card surface (pure white).
- **Card Foreground (`oklch(0.12 0.01 240)`):** Card text.
- **Popover (`oklch(1 0 0)`):** Popover surface.
- **Popover Foreground (`oklch(0.12 0.01 240)`):** Popover text.
- **Primary (`oklch(0.25 0.02 240)`):** Primary action (near-black).
- **Primary Foreground (`oklch(0.99 0.003 240)`):** Text on primary.
- **Secondary (`oklch(0.95 0.005 240)`):** Secondary surface.
- **Secondary Foreground (`oklch(0.25 0.02 240)`):** Text on secondary.
- **Muted (`oklch(0.95 0.005 240)`):** Muted background.
- **Muted Foreground (`oklch(0.45 0.015 240)`):** Muted/secondary text.
- **Accent (`oklch(0.95 0.005 240)`):** Accent surface (hover states).
- **Accent Foreground (`oklch(0.25 0.02 240)`):** Text on accent.
- **Destructive (`oklch(0.55 0.22 25)`):** Destructive actions (red).
- **Border (`oklch(0.9 0.005 240)`):** Border color.
- **Input (`oklch(0.9 0.005 240)`):** Input border.
- **Ring (`oklch(0.6 0.02 240)`):** Focus ring.

### Dark Mode (`.dark`)

- **Background (`oklch(0.08 0.005 240)`):** Near-black.
- **Foreground (`oklch(0.98 0.005 240)`):** Near-white.
- **Card (`oklch(0.12 0.008 240)`):** Elevated surface.
- **Card Foreground (`oklch(0.98 0.005 240)`):** Card text.
- **Popover (`oklch(0.12 0.008 240)`):** Same as card.
- **Popover Foreground (`oklch(0.98 0.005 240)`):** Popover text.
- **Primary (`oklch(0.92 0.005 240)`):** Inverted (near-white).
- **Primary Foreground (`oklch(0.12 0.008 240)`):** Dark text on light primary.
- **Secondary (`oklch(0.18 0.01 240)`):** Secondary surface.
- **Secondary Foreground (`oklch(0.98 0.005 240)`):** Text on secondary.
- **Muted (`oklch(0.18 0.01 240)`):** Muted background.
- **Muted Foreground (`oklch(0.65 0.01 240)`):** Muted text.
- **Accent (`oklch(0.18 0.01 240)`):** Accent surface.
- **Accent Foreground (`oklch(0.98 0.005 240)`):** Text on accent.
- **Destructive (`oklch(0.6 0.18 25)`):** Slightly brighter red.
- **Border (`oklch(0.2 0.01 240 / 40%)`):** Border with alpha transparency.
- **Input (`oklch(0.2 0.01 240 / 50%)`):** Input with alpha transparency.
- **Ring (`oklch(0.5 0.02 240)`):** Focus ring.

### Sidebar Tokens

The sidebar has a dedicated token set mirroring the core palette. In light mode it matches `--background`; in dark mode it matches `--card`.

| Token | Light | Dark |
|---|---|---|
| `--sidebar` | `oklch(0.99 0.003 240)` | `oklch(0.12 0.008 240)` |
| `--sidebar-foreground` | `oklch(0.12 0.01 240)` | `oklch(0.98 0.005 240)` |
| `--sidebar-primary` | `oklch(0.25 0.02 240)` | `oklch(0.92 0.005 240)` |
| `--sidebar-primary-foreground` | `oklch(0.99 0.003 240)` | `oklch(0.12 0.008 240)` |
| `--sidebar-accent` | `oklch(0.95 0.005 240)` | `oklch(0.18 0.01 240)` |
| `--sidebar-accent-foreground` | `oklch(0.25 0.02 240)` | `oklch(0.98 0.005 240)` |
| `--sidebar-border` | `oklch(0.9 0.005 240)` | `oklch(0.2 0.01 240 / 40%)` |
| `--sidebar-ring` | `oklch(0.6 0.02 240)` | `oklch(0.5 0.02 240)` |

### Chart Tokens

Used for data visualization. Identical in both themes:

| Token | Value |
|---|---|
| `--chart-1` | `oklch(0.646 0.222 41.116)` |
| `--chart-2` | `oklch(0.6 0.118 184.704)` |
| `--chart-3` | `oklch(0.398 0.07 227.392)` |
| `--chart-4` | `oklch(0.828 0.189 84.429)` |
| `--chart-5` | `oklch(0.769 0.188 70.08)` |

## 3. Typography Rules
* **Display & Buttons:** `Geist Mono` (or fallback monospace) is used for headings, metrics, IDs, hashes, and button text. Headings use slightly negative tracking (`-0.02em`), while buttons and tags use uppercase with positive letter-spacing (`0.06em`) to prevent crowding.
* **Body & UI Text:** `Geist Sans` (or fallback Inter/system-ui) is used for body and general UI text. Body text is set to `1.6` line-height for enhanced readability.

### Font Stack

- Sans-serif: **Geist Sans** (`--font-geist-sans`) — loaded via `next/font/google`
- Monospace: **Geist Mono** (`--font-geist-mono`) — loaded via `next/font/google`

### OpenType Features

```css
font-feature-settings: "cv02", "cv03", "cv04", "cv11";
```

### Text Rendering

- Body: `antialiased` (via Tailwind class on `<body>`)
- Tabular numbers: `.tabular-nums { font-variant-numeric: tabular-nums; }`

### Observed Type Scale (from component usage)

| Context | Classes | Notes |
|---|---|---|
| Hero heading | `text-6xl font-extrabold tracking-tighter sm:text-7xl md:text-8xl` | Landing page |
| Page title (Card) | `text-xl font-semibold` | Settings, dashboard headings |
| Card title | `text-base font-medium leading-snug` | Default CardTitle |
| Dialog title | `text-lg leading-none font-semibold` | — |
| Body / description | `text-sm text-muted-foreground` | CardDescription, FieldDescription |
| Input text | `text-base md:text-sm` | Responsive sizing |
| Button text | `text-sm font-medium` | All button variants |
| Field label | `text-sm font-medium` | Via Label component |
| Mono / code | `font-mono text-xs` | Welcome-back email, code blocks |
| Error text | `text-sm font-normal text-destructive` | FieldError |

## 4. Component Stylings
* **Buttons:** Fully sharp edges (`--radius-button: 0px`). All buttons use `@base-ui/react/button` as the underlying primitive. The `destructive` variant is now a soft tint (`bg-destructive/10 text-destructive`), not a filled background — reserving solid destructive fills for `AlertDialog` confirm actions only.
* **Cards/Containers:** Sharp 4px radius (`--radius-card: 4px`). Uses `ring-1 ring-foreground/10` for elevation instead of `border shadow-sm`. Absolutely zero box-shadows are used for standard cards.
* **Inputs/Forms:** Built on `@base-ui/react/input`. Defined by `border border-input` stroke and transparent background. Focus states use `focus-visible:ring-3 focus-visible:ring-ring/50`. No rounded pills or soft shadow effects.

### UI Primitives (`@hasir/ui/components/`)

#### Button

CVA-based with 6 variants and 7 sizes. Built on `@base-ui/react/button`. Custom `isLoading` prop adds a `Spinner` and disables interaction.

**Variants:**

| Variant | Light Appearance | Dark Adjustment |
|---|---|---|
| `default` | `bg-primary text-primary-foreground hover:bg-primary/80` | Primary inverts (near-white bg) |
| `destructive` | `bg-destructive/10 text-destructive hover:bg-destructive/20` | `dark:bg-destructive/20 dark:hover:bg-destructive/30` |
| `outline` | `border-border bg-background hover:bg-muted` | `dark:border-input dark:bg-input/30 dark:hover:bg-input/50` |
| `secondary` | `bg-secondary text-secondary-foreground hover:bg-[color-mix(...)]` | — |
| `ghost` | `hover:bg-muted hover:text-foreground` | `dark:hover:bg-muted/50` |
| `link` | `text-primary underline-offset-4 hover:underline` | — |

**Sizes:**

| Size | Dimensions | Notes |
|---|---|---|
| `default` | `h-8 gap-1.5 px-2.5` | Standard action buttons |
| `xs` | `h-6 px-2 text-xs` | Compact inline actions; uses `rounded-md` override |
| `sm` | `h-7 px-2.5 text-[0.8rem]` | Small actions; uses `rounded-md` override |
| `lg` | `h-9 px-2.5` | Larger emphasis buttons |
| `icon` | `size-8` (32px square) | Standard icon button |
| `icon-xs` | `size-6` (24px square) | Small icon button |
| `icon-sm` | `size-7` (28px square) | Medium icon button |
| `icon-lg` | `size-9` (36px square) | Large icon button |

> [!NOTE]
> `xs` and `sm` sizes use `rounded-[min(var(--radius-md),12px)]` — they have slightly rounded corners to feel proportional despite `--radius-button: 0px` on default/lg sizes.

**Loading State:**
- Prepends an animated `Spinner` (`Loader2Icon` with `animate-spin`)
- Sets `disabled` automatically when `isLoading` is true
- SVGs inside buttons default to `size-4` via `[&_svg:not([class*='size-'])]:size-4`

**Focus Style:** `focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50`

**Aria support:** `aria-expanded` adjusts background; `aria-invalid` adds destructive ring; `active:not-aria-[haspopup]:translate-y-px` for press feedback.

#### Card

Composable sub-components: `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardAction`, `CardContent`, `CardFooter`.

Has a `size` prop (`"default"` | `"sm"`) that scales `--card-spacing` from `--spacing(4)` (16px) to `--spacing(3)` (12px).

| Part | Key Classes |
|---|---|
| Card | `flex flex-col rounded-[var(--radius-card)] bg-card ring-1 ring-foreground/10 py-(--card-spacing) gap-(--card-spacing)` |
| CardHeader | `px-(--card-spacing) gap-1`, grid with `auto-rows-min`, action slot support |
| CardTitle | `text-base font-medium leading-snug` (sm: `text-sm`) |
| CardDescription | `text-sm text-muted-foreground` |
| CardContent | `px-(--card-spacing)` |
| CardFooter | `flex items-center rounded-b-[var(--radius-card)] border-t bg-muted/50 p-(--card-spacing)` |

All parts use `data-slot` attributes for CSS targeting.

> [!IMPORTANT]
> Card no longer uses `border shadow-sm`. Elevation is achieved via `ring-1 ring-foreground/10`, which renders as a hairline outline that works in both light and dark modes without alpha-blending issues.

#### Input

Built on `@base-ui/react/input`. No longer a masked password variant — inputs are standard native inputs.

**Base Input Style:**
```css
h-8 w-full rounded-lg border border-input bg-transparent px-2.5 py-1
text-base md:text-sm transition-colors outline-none
dark:bg-input/30
```

**Focus:** `focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50`
**Invalid:** `aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40`
**Disabled:** `disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50`

#### Alert

CVA with 2 variants: `default` and `destructive`.

| Variant | Style |
|---|---|
| `default` | `bg-card text-card-foreground` |
| `destructive` | `text-destructive bg-card [&>svg]:text-current` |

Grid layout: `grid-cols-[calc(var(--spacing)*4)_1fr]` when an SVG icon is present; otherwise single-column.

#### Dialog

Radix-based. Content centered with `fixed top-[50%] left-[50%] translate-x-[-50%] translate-y-[-50%]`.

- Max width: `max-w-[calc(100%-2rem)] sm:max-w-lg`
- Overlay: `bg-black/50`
- Animation: fade-in/out + zoom 95%→100%
- Close button: top-right `XIcon`, optional via `showCloseButton` prop
- Footer: `flex-col-reverse gap-2 sm:flex-row sm:justify-end`

#### AlertDialog

Same visual treatment as Dialog but without dismissible close button. Uses `AlertDialogAction` (primary variant) and `AlertDialogCancel` (outline variant) from `buttonVariants`. Reserved for **destructive confirmations** — delete account, delete organization, remove member.

#### Field System

Composable form field primitives: `FieldSet`, `FieldGroup`, `Field`, `FieldLabel`, `FieldTitle`, `FieldDescription`, `FieldError`, `FieldContent`, `FieldSeparator`.

**Field Orientations (CVA):**

| Orientation | Layout |
|---|---|
| `vertical` (default) | `flex-col`, children full-width |
| `horizontal` | `flex-row items-center`, label auto-fills |
| `responsive` | Vertical by default, horizontal at `@md` container query |

**FieldError:** Renders role="alert", deduplicates error messages, shows as bullet list when multiple.
**FieldDescription:** Links inside descriptions get `underline underline-offset-4 hover:text-primary`.

#### Empty

Composable empty-state primitives: `Empty`, `EmptyHeader`, `EmptyMedia`, `EmptyTitle`, `EmptyDescription`, `EmptyContent`.

| Part | Key Classes |
|---|---|
| `Empty` | `flex flex-col items-center justify-center gap-4 rounded-xl border-dashed p-6 text-center` |
| `EmptyHeader` | `flex max-w-sm flex-col items-center gap-2` |
| `EmptyMedia` | CVA: `default` (transparent bg) or `icon` (muted bg square, `size-8`) |
| `EmptyTitle` | `text-sm font-medium tracking-tight` |
| `EmptyDescription` | `text-sm/relaxed text-muted-foreground` with auto-linked anchors |
| `EmptyContent` | `flex flex-col items-center gap-2.5` for CTA buttons |

#### Pagination

Custom component (not shadcn default). Uses outline buttons with delta-2 sliding window.

- Prev/Next: `ChevronLeft`/`ChevronRight` icons, outline variant, sm size, `h-8 w-8 p-0`
- Active page: `default` variant; inactive: `outline` variant
- Ellipsis: `MoreHorizontal` icon in `text-muted-foreground`
- Hidden when `totalPages <= 1`

#### Spinner

Wrapper around `Loader2Icon` with `animate-spin`, `size-4` default. Has `role="status"` and `aria-label="Loading"`.

#### Sonner (Toaster)

Theme-aware toast system with custom Lucide icons:

| Level | Icon |
|---|---|
| Success | `CircleCheckIcon` |
| Info | `InfoIcon` |
| Warning | `TriangleAlertIcon` |
| Error | `OctagonXIcon` |
| Loading | `Loader2Icon` (animated) |

All icons are `size-4`. Styling uses design tokens:
```css
--normal-bg: var(--popover);
--normal-text: var(--popover-foreground);
--normal-border: var(--border);
--border-radius: var(--radius);
```

### Feature Components

- **Authentication Forms:** `login-form`, `register-form`, `forgot-password-form`, `reset-password-form`, `password-confirmation-dialog`. Follow centered card layout, zod validation, explicit loading/error states.
- **Header:** Fixed top bar with brand link, quick search (`Cmd+K`) that opens a live `SearchDropdown`, creation dropdown, theme toggler, and user menu.
- **Search Dropdown:** `search-dropdown` — a `forwardRef` dropdown surfacing live Organization and Repository results with skeleton loading states.
- **Dashboard:** Listing cards for Organizations and Repositories with pagination and `Empty`-component empty states.
- **Organization Detail:** Uses a vertical sidebar tabs layout with sub-routes for Users, Repositories, and Settings.
- **Repository Detail:** Vertical sidebar tabs layout with sub-routes for Documentation, Files, Commits, SDK Preferences, and Settings.
- **Member Management:** `members-list`, `member-item`, `invite-user-dialog`, `delete-member-dialog`. Permission-gated based on roles.
- **Organization Management:** `organization-settings-content`, `organization-settings-form`, `organization-repositories-content`, `organization-users-content`, `delete-organization-dialog`.
- **Repository Management:** `repository-settings-content`, `repository-settings-form`, `repository-files-content`, `repository-commits-content`, `repository-documentation-content`, `repository-sdk-preferences-content`, `delete-repository-dialog`.
- **SDK URLs:** `sdk-urls` — dropdown with copy-to-clipboard for SDK access URLs per language/version.
- **SSH & API Key Panel:** `ssh-api-key-panel` — manages SSH public keys and API keys (list, add, delete, copy) within a card using pagination.
- **SSH Configuration Docs:** Public docs page at `/docs/ssh-configuration` with `ssh-configuration-content` and `ssh-configuration-client`.
- **Danger Zone Pattern:** A card with `border-destructive/50` containing a destructive action button that triggers a confirmation Dialog.
- **Clone URLs:** Popover with SSH/HTTPS clone links and copy-to-clipboard actions.
- **Markdown Renderer:** Built with `react-markdown` + `remark-gfm` + `rehype-raw` + `rehype-sanitize`.
- **SDK Install Guide Dialog:** Dynamic guides showing language-specific SDK install commands and syntax-highlighted setup code.

### Animation & Motion

#### Tailwind Animations (tw-animate-css)
Used for Dialog/AlertDialog/Popover/DropdownMenu enter/exit:
- `animate-in` / `animate-out`
- `fade-in-0` / `fade-out-0`
- `zoom-in-95` / `zoom-out-95`

#### Framer Motion
Used exclusively on the landing page (`home-page-content.tsx`):
- **Container pattern:** staggered children fade-in.
- **Item pattern:** spring physics y-axis transition.

#### CSS Transitions
- Buttons: `transition-all` base, active translate-y press transitions.
- Theme toggle icons: rotation + scale transitions.
- Hover states: `transition-colors duration-500` (e.g. landing cards).

## 5. Layout Principles
* **Whitespace & Grids:** Anchored to an 8px baseline grid with clean alignment and generous margins.
* **Aesthetic Restraint:** Structure is defined purely by `ring-1 ring-foreground/10` hairline outlines rather than shadows or borders. The amber accent is strictly reserved for high-signal alerts or cursors rather than general styling.

### Layout & Spacing conventions

| Context | Gap / Padding |
|---|---|
| Card internal gap | `gap-(--card-spacing)` = `gap-4` (16px) default, `gap-3` (12px) sm |
| Card padding | `px-(--card-spacing) py-(--card-spacing)` |
| Form field gap | `gap-6` via FieldSet, `gap-7` via FieldGroup |
| Dialog padding | `p-6` with `gap-4` between sections |
| Section spacing | `gap-4` between form fields, `gap-1.5` between label+input |
| Page max-width | `max-w-2xl mx-auto` (typical content pages) |
| Button internal gap | `gap-1.5` (default/lg), `gap-1` (sm/xs) |

### Provider Stack

```
<TransportProvider>          ← ConnectRPC transport (binary, with csrf + auth + idempotency interceptors)
  <QueryClientProvider>      ← TanStack Query cache
    <SessionProvider>        ← JWT session context
      <ThemeProvider>        ← next-themes (class strategy)
        <main>{children}</main>
        <Toaster />          ← Sonner toast container
      </ThemeProvider>
    </SessionProvider>
  </QueryClientProvider>
</TransportProvider>
```

**ConnectRPC Interceptors (applied in order):**
1. `csrfInterceptor` — injects CSRF token from session cookie into every mutating request
2. `authInterceptor` — attaches Bearer JWT from iron-session to Authorization header
3. `idempotencyInterceptor` — adds `Idempotency-Key` header to safe non-idempotent mutations

### Route Structure

```
app/
├── layout.tsx                    # Root: Geist fonts, antialiased body, Providers
├── providers.tsx                 # Client-side: Transport + Query + Session + Theme + Toaster
├── page.tsx                      # Landing: HomePageContent (Framer Motion)
├── login/page.tsx                # Public: LoginForm
├── register/page.tsx             # Public: RegisterForm
├── forgot-password/page.tsx      # Public: ForgotPasswordForm
├── reset-password/[token]/page.tsx # Public: ResetPasswordForm
├── docs/
│   └── ssh-configuration/page.tsx # Public: SSH setup guide (no header)
├── api/
│   ├── auth/login/               # iron-session login route
│   ├── auth/logout/              # iron-session logout route
│   ├── auth/session/             # JWT session read route
│   └── docs/[organizationId]/[repositoryId]/[commitHash]/
│                                 # Server-side proxy: fetches proto docs from API
└── (authenticated)/
    ├── layout.tsx                # HeaderClient wrapper
    ├── invite/[token]/page.tsx   # InviteResponse (within auth group)
    ├── dashboard/page.tsx        # Dashboard (org/repo listing)
    ├── profile/page.tsx          # ProfilePageContent (tabs)
    ├── organization/
    │   └── [id]/
    │       ├── layout.tsx        # Sidebar layout (tabs)
    │       ├── page.tsx          # Overview redirect
    │       ├── users/page.tsx    # Member management
    │       ├── repositories/page.tsx # Org repositories
    │       └── settings/page.tsx # Org settings + danger zone
    └── repository/
        └── [repositoryId]/
            ├── layout.tsx        # Sidebar layout (tabs)
            ├── page.tsx          # Overview redirect
            ├── documentation/page.tsx
            ├── files/page.tsx
            ├── commits/page.tsx
            ├── sdk-preferences/page.tsx
            └── settings/page.tsx
```

### Page Layout Patterns

- **Landing page:** Full-viewport hero, centered content, `noise-bg` texture, ambient spotlights, Framer Motion stagger animations.
- **Auth pages:** Centered `Card` at `max-w-md`, minimal chrome, no header.
- **Docs pages:** Public, no header, standalone layout (e.g. SSH configuration guide).
- **Authenticated pages:** `HeaderClient` at top, content below. No fixed sidebar at the app level — sidebars are per-entity (org, repo).
- **Entity detail pages (org, repo):** Two-column layout — left sidebar nav (tabs) + right content area. Sidebar rendered as vertical `Tabs` list with Radix.
- **Settings / profile pages:** Single-column centered content at `max-w-2xl`, cards stacked vertically.

## 6. Elevation & Depth

Visual depth is achieved through **Tonal Layers** rather than heavy shadows. The background uses a soft off-white or very light grey, while primary content sits on pure white cards outlined by `ring-1 ring-foreground/10`.

### Premium Visual Texture

The landing page uses a subtle visual texture system:

```css
/* Noise overlay — 2.5% opacity, fixed, non-interactive */
.noise-bg::before {
  content: "";
  position: fixed;
  inset: 0;
  opacity: 0.025;
  z-index: 9999;
  pointer-events: none;
  background-image: url("data:image/svg+xml,...feTurbulence...");
}

/* Glassmorphism card */
.premium-glass {
  background: rgba(var(--card), 0.7);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.03);
}

/* Dark mode glass */
.dark .premium-glass {
  background: rgba(var(--card), 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.2);
}
```

The landing page also uses ambient spotlight blurs:
```css
bg-primary/5 rounded-full blur-[120px]  /* top-left */
bg-secondary/5 rounded-full blur-[140px] /* bottom-right */
```

## 7. Shapes

Base radius: `--radius: 0.75rem` (12px). Computed variants:

| Token | Computation | Result |
|---|---|---|
| `--radius-sm` | `calc(var(--radius) - 4px)` | 8px |
| `--radius-md` | `calc(var(--radius) - 2px)` | 10px |
| `--radius-lg` | `var(--radius)` | 12px |
| `--radius-xl` | `calc(var(--radius) + 4px)` | 16px |
| `--radius-card` | Fixed | **4px** — sharp card corners |
| `--radius-button` | Fixed | **0px** — fully flat button corners |

Components use `rounded-md` (10px) by default in popovers/dropdowns. Cards use `rounded-[var(--radius-card)]` (4px). Dialogs use `rounded-lg` (12px). Buttons are fully sharp (`0px`) except `xs`/`sm` sizes which use `rounded-[min(var(--radius-md),12px)]` for proportionality.

## 8. Do's and Don'ts

- **Do** use OKLCH color space for any new color tokens.
- **Do** use `cn()` for all class merging — never concatenate raw class strings.
- **Do** use `cva` for components requiring variant styling.
- **Do** include `data-slot` attributes on all custom UI primitives.
- **Do** use `react-hook-form` + `zod` for all form modules.
- **Do** use `sonner` toasts for transient operation feedbacks.
- **Do** use `Dialog` for creation and input interfaces, reserving `AlertDialog` for confirmation steps.
- **Do** use the `isLoading` prop on buttons to automatically handle disabling and spinning indicators.
- **Do** use Framer Motion solely for page-level entrance animations.
- **Do** use Lucide icons sized to `size-4` by default with `pointer-events-none`.
- **Do** use a vertical sidebar tab structure for entity details (org/repo dashboards).
- **Do** style danger zones with `border-destructive/50`.
- **Do** manage accessibility roles (`role="alert"`, `role="status"`, `aria-busy`, etc.) across all interactions.
- **Do** use `Empty` / `EmptyHeader` / `EmptyMedia` / `EmptyTitle` / `EmptyDescription` / `EmptyContent` for empty states — never roll custom empty state markup.
- **Do** use `@base-ui/react/button` and `@base-ui/react/input` primitives via the `@hasir/ui` wrappers — never import from `@base-ui/react` directly in the dashboard.
- **Don't** suppress type checking errors with `as any`, `@ts-ignore`, or `@ts-expect-error`.
- **Don't** use inline alerts for transient success/error messages.
- **Don't** mix rounded and sharp corners in the same view.
- **Don't** define custom spacing classes; always stick to Tailwind defaults.
- **Don't** hardcode light/dark theme values; always refer to the CSS theme variables.
- **Don't** add `border` or `shadow` to cards — use `ring-1 ring-foreground/10` for the hairline outline.
- **Don't** use the `destructive` button variant for filled red CTAs — it's a soft tint. Use `AlertDialog` with its `AlertDialogAction` for confirmed destructive operations.
