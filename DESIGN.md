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
    fontWeight: 600
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

spacing:
  # Dashboard Spacing
  xs: 4px
  sm: 8px
  md: 16px
  lg: 32px
  xl: 64px
  gap-card: 24px
  padding-card: 24px

components:
  button-default:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
    padding: 16px
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "#ffffff"
    rounded: "{rounded.md}"
    padding: 16px
  button-outline:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: 16px
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.secondary-foreground}"
    rounded: "{rounded.md}"
    padding: 16px
  button-ghost:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.accent-foreground}"
    rounded: "{rounded.md}"
    padding: 16px
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    rounded: "{rounded.xl}"
    padding: 24px
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
    backgroundColor: "{colors.input}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
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
| Framework | Next.js (App Router) | 16.2.10 |
| UI Library | React | 19.2.7 |
| Language | TypeScript | 6.x (strict) |
| Package Manager | Bun | ~1.3 |
| Styling | Tailwind CSS | 4.x |
| Component System | shadcn/ui (new-york style) | 4.x |
| Primitives | Radix UI | Various |
| Variant Management | class-variance-authority (CVA) | 0.7.x |
| Class Merging | clsx + tailwind-merge via `cn()` | — |
| Icons | Lucide React | 1.24.x |
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

### Configuration

**shadcn/ui** (`components.json`):

```json
{
  "style": "new-york",
  "rsc": true,
  "tsx": true,
  "tailwind": {
    "css": "app/globals.css",
    "baseColor": "neutral",
    "cssVariables": true
  },
  "iconLibrary": "lucide",
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  },
  "registries": {
    "@magicui": "https://magicui.design/r/{name}.json"
  }
}
```

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
| Card title | `font-semibold leading-none` | Default CardTitle |
| Dialog title | `text-lg leading-none font-semibold` | — |
| Body / description | `text-sm text-muted-foreground` | CardDescription, FieldDescription |
| Input text | `text-base md:text-sm` | Responsive sizing |
| Button text | `text-sm font-medium` | All button variants |
| Field label | `text-sm font-medium` | Via Label component |
| Mono / code | `font-mono text-xs` | Welcome-back email, code blocks |
| Error text | `text-sm font-normal text-destructive` | FieldError |

## 4. Component Stylings
* **Buttons:** Sharp, squared-off edges (0px border radius). Primary CTAs use dark text on a white filled background. Active and hover states employ high-contrast monochrome transitions.
* **Cards/Containers:** Subtly rounded corners (4px radius). Styled with Surface Panel background (`#101213`) and Hairline Border (`#202223`). Absolutely zero box-shadows are used for elevation.
* **Inputs/Forms:** Defined by Hairline Border stroke and Surface Panel background. Focus states utilize a precise White Signal or Warm Amber Accent hairline border. No rounded pills or soft shadow effects.

### UI Primitives (`components/ui/`)

#### Button

CVA-based with 6 variants and 6 sizes. Custom `isLoading` prop adds a `Spinner` and disables interaction.

**Variants:**

| Variant | Light Appearance | Dark Adjustment |
|---|---|---|
| `default` | `bg-primary text-primary-foreground hover:bg-primary/90` | Primary inverts (near-white bg) |
| `destructive` | `bg-destructive text-white hover:bg-destructive/90` | `dark:bg-destructive/60` |
| `outline` | `border bg-background shadow-xs hover:bg-accent` | `dark:bg-input/30 dark:border-input dark:hover:bg-input/50` |
| `secondary` | `bg-secondary text-secondary-foreground hover:bg-secondary/80` | — |
| `ghost` | `hover:bg-accent hover:text-accent-foreground` | `dark:hover:bg-accent/50` |
| `link` | `text-primary underline-offset-4 hover:underline` | — |

**Sizes:**

| Size | Dimensions |
|---|---|
| `default` | `h-9 px-4 py-2` (with icon: `px-3`) |
| `sm` | `h-8 px-3` (with icon: `px-2.5`, `gap-1.5`) |
| `lg` | `h-10 px-6` (with icon: `px-4`) |
| `icon` | `size-9` (36px square) |
| `icon-sm` | `size-8` (32px square) |
| `icon-lg` | `size-10` (40px square) |

**Loading State:**
- Prepends an animated `Spinner` (`Loader2Icon` with `animate-spin`)
- Sets `cursor-progress`, `disabled`, `aria-busy`, and `data-loading`
- SVGs inside buttons default to `size-4` via `[&_svg:not([class*='size-'])]:size-4`

**Focus Style:** `focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]`

**Pressed Animation:** `active:scale-[0.98] active:translate-y-[1px] transition-transform duration-100`

#### Card

Composable sub-components: `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardAction`, `CardContent`, `CardFooter`.

| Part | Key Classes |
|---|---|
| Card | `bg-card text-card-foreground rounded-xl border shadow-sm py-6 gap-6` |
| CardHeader | `px-6 gap-2`, grid with auto-rows, action slot support |
| CardTitle | `font-semibold leading-none` |
| CardDescription | `text-muted-foreground text-sm` |
| CardContent | `px-6` |
| CardFooter | `flex items-center px-6` |

All parts use `data-slot` attributes for CSS targeting.

#### Input

Custom implementation with a **masked password** variant. When `type="password"`, the Input component renders a `PasswordInput` that replaces each character with `*` while maintaining the actual value in a ref. This prevents browser autofill preview leaks.

**Base Input Style:**
```css
h-9 rounded-md border bg-transparent px-3 py-1 text-base md:text-sm shadow-xs
selection:bg-primary selection:text-primary-foreground
dark:bg-input/30
```

**Focus:** `focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]`
**Invalid:** `aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive`

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
- **Header:** Fixed top bar with brand link, quick search (`Cmd+K`), creation dropdown, theme toggler, and user menu.
- **Dashboard:** Listing cards for Organizations and Repositories with pagination and custom empty states.
- **Organization / Repository Detail:** Uses a vertical sidebar tabs layout (`Radix Tab` component) for navigation.
- **Member Management:** `members-list`, `member-item`, `invite-user-dialog`, `invite-response`. Permission gated based on roles.
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
- Buttons: `transition-all` base, active scale click transitions.
- Theme toggle icons: rotation + scale transitions.
- Hover states: `transition-colors duration-500` (e.g. landing cards).

## 5. Layout Principles
* **Whitespace & Grids:** Anchored to an 8px baseline grid with clean alignment and generous margins.
* **Aesthetic Restraint:** Structure is defined purely by 1px hairline borders rather than shadows. The amber accent is strictly reserved for high-signal alerts or cursors rather than general styling.

### Layout & Spacing conventions

| Context | Gap / Padding |
|---|---|
| Card internal gap | `gap-6` (24px) between header/content/footer |
| Card padding | `px-6 py-6` (24px) |
| Form field gap | `gap-6` via FieldSet, `gap-7` via FieldGroup |
| Dialog padding | `p-6` with `gap-4` between sections |
| Section spacing | `gap-4` between form fields, `gap-1.5` between label+input |
| Page max-width | `max-w-2xl mx-auto` (typical content pages) |
| Button internal gap | `gap-2` (default), `gap-1.5` (sm) |

### Provider Stack

```
<TransportProvider>          ← ConnectRPC transport (binary, with auth interceptor)
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

### Route Structure

```
app/
├── layout.tsx               # Root: Geist fonts, antialiased body, Providers
├── page.tsx                 # Landing: HomePageContent (Framer Motion)
├── login/page.tsx           # Public: LoginForm
├── register/page.tsx        # Public: RegisterForm
├── forgot-password/page.tsx # Public: ForgotPasswordForm
├── reset-password/page.tsx  # Public: ResetPasswordForm
├── invite/[token]/page.tsx  # Public: InviteResponse
└── (authenticated)/
    ├── layout.tsx           # HeaderClient wrapper
    ├── dashboard/page.tsx   # Dashboard (org/repo listing)
    ├── profile/page.tsx     # ProfilePageContent (tabs)
    ├── organization/
    │   └── [id]/
    │       ├── layout.tsx   # Sidebar layout (Users/Org/Repos tabs)
    │       └── page.tsx
    └── repository/
        └── [repositoryId]/
            ├── layout.tsx   # Sidebar layout (Docs/Files/Commits/SDK/Settings)
            └── page.tsx
```

### Page Layout Patterns

- **Landing page:** Full-viewport hero, centered content, `noise-bg` texture, ambient spotlights, Framer Motion stagger animations.
- **Auth pages:** Centered `Card` at `max-w-md`, minimal chrome, no header.
- **Authenticated pages:** `HeaderClient` at top, content below. No fixed sidebar at the app level — sidebars are per-entity (org, repo).
- **Entity detail pages (org, repo):** Two-column layout — left sidebar nav (tabs) + right content area. Sidebar rendered as vertical `Tabs` list with Radix.
- **Settings / profile pages:** Single-column centered content at `max-w-2xl`, cards stacked vertically.

## 6. Elevation & Depth

Visual depth is achieved through **Tonal Layers** rather than heavy shadows. The background uses a soft off-white or very light green, while primary content sits on pure white cards.

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

Components use `rounded-md` (10px) by default. Cards use `rounded-xl` (16px). Dialogs use `rounded-lg` (12px).

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
- **Don't** suppress type checking errors with `as any`, `@ts-ignore`, or `@ts-expect-error`.
- **Don't** use inline alerts for transient success/error messages.
- **Don't** mix rounded and sharp corners in the same view.
- **Don't** define custom spacing classes; always stick to Tailwind defaults.
- **Don't** hardcode light/dark theme values; always refer to the CSS theme variables.