---
name: Hybrid PQS PoC
description: Post-quantum Ethereum transaction signing toolkit.
colors:
  paper-white: "oklch(1 0 0)"
  carbon: "oklch(0.129 0.042 264.695)"
  charcoal: "oklch(0.208 0.042 265.755)"
  fog: "oklch(0.929 0.013 255.508)"
  ash: "oklch(0.704 0.04 256.788)"
  silver: "oklch(0.968 0.007 247.896)"
  warm-gray: "oklch(0.554 0.046 257.417)"
  signal-red: "oklch(0.577 0.245 27.325)"
typography:
  title:
    fontFamily: "Inter, sans-serif"
    fontSize: "1rem"
    fontWeight: 500
    lineHeight: 1.5
  body:
    fontFamily: "Inter, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Inter, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    letterSpacing: "0.02em"
rounded:
  sm: "calc(0.625rem - 4px)"
  md: "calc(0.625rem - 2px)"
  lg: "0.625rem"
spacing:
  xs: "0.25rem"
  sm: "0.5rem"
  md: "1rem"
  lg: "2rem"
  xl: "3rem"
components:
  button-primary:
    backgroundColor: "{colors.charcoal}"
    textColor: "{colors.paper-white}"
    rounded: "{rounded.lg}"
    padding: "0.5rem 1rem"
  button-primary-hover:
    backgroundColor: "oklch(0.16 0.038 266)"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.carbon}"
    rounded: "{rounded.lg}"
  input-default:
    borderColor: "{colors.fog}"
    backgroundColor: "transparent"
    rounded: "{rounded.lg}"
---

# Design System: Hybrid PQS PoC

**Creative North Star: "The Cryptographic Laboratory"**

A zero-decorative research surface where the interface behaves like a calibrated instrument panel. Color stays neutral so algorithm parameters, private keys, and signature results read first; the only saturated color appears on actions that alter on-chain state (red) or confirm one (charcoal). Typography is single-family (Inter) with weight and scale doing the hierarchy, not novelty. Surfaces rest flat; the sole elevation is a 1px shadow that lifts floating task cards off the page field.

The user is a researcher running an experiment: generate keypair, fund the ledger, submit a signed transaction, verify it. Every control is a verified input, not decoration.

## Colors

Monochrome discipline with a single functional accent. Light theme carries the full set below; a parallel dark set (`@media (prefers-color-scheme: dark)`-aware `.dark`) inverts background/foreground and deepens card surfaces.

### Primary
- **Charcoal** (`oklch(0.208 0.042 265.755)`): the action color — primary buttons, badges, connected-state labels. A neutral dark gray, deliberately chosen over a brand blue to keep the interface feeling like lab equipment rather than a SaaS product.

### Secondary
- **Silver** (`oklch(0.968 0.007 247.896)`): secondary/ghost buttons, `secondary`/`muted`/`accent` tokens. The quiet companion to Charcoal.

### Tertiary (Accent)
- **Signal Red** (`oklch(0.577 0.245 27.325)`): reserved exclusively for destructive/interruptive actions (withdraw, delete). Its rarity is the point.

### Neutral
- **Paper White** (`oklch(1 0 0)`): page field and card surfaces; also `primary-foreground`.
- **Carbon** (`oklch(0.129 0.042 264.695)`): primary text and icon color (`foreground`).
- **Fog** (`oklch(0.929 0.013 255.508)`): borders and input strokes (`border`/`input`).
- **Ash** (`oklch(0.704 0.04 256.788)`): focus ring (`ring`) and subtle divider.
- **Warm Gray** (`oklch(0.554 0.046 257.417)`): secondary text (`muted-foreground`).

**The Single Red Rule.** Signal Red appears only on actions that move on-chain funds or delete state. It is never used for primary/affirmative actions or links.

## Typography

**Body Font (and only font):** Inter, with a system `sans-serif` fallback.

### Hierarchy
- **Title (500/1rem):** card headers and section titles (e.g. "Send Transaction", "Wallet Info").
- **Body (400/0.875rem):** form labels, balance lines, descriptive copy.
- **Label (500/0.75rem/0.02em tracking):** helper text and inline hints (e.g. "Expects a concatenated Hex key…").

No decorative display. Size variance is small (14px→16px) to keep the panel dense without sacrificing legibility at a glance.

## Layout

Single-center layout: a `max-w-5xl` viewport, horizontally centered, with `px-8 py-5` padding above the fold and `gap-8` between major blocks. Inside cards, vertical rhythm is `gap-6` (form row) / `gap-4` (field group) / `gap-3` (tight pairs). The grid splits `md:grid-cols-2` when two panels fit (wallet info + action form). No responsive breakpoints beyond the md split; density is comfortable, not compact.

## Elevation & Depth

Flat-by-default with a single soft lift. Floating task cards carry Tailwind `shadow-sm` (a 1px surface-shadow only); the header sits flush with a `border-b` and no shadow. Depth is conveyed by border + that one shadow, never by layered cards or ambient glow.

## Shapes

- **Form fields, cards, buttons:** `rounded-md` (mapped to the theme's `--radius-md` ≈ 8px).
- **Badges:** `rounded-full`.
- **Border weight:** 1px strokes throughout (`border`, `border-input`). No heavy frames.

## Components

### Buttons
- **Shape:** `rounded-md` (~8px), `h-9`, inline-flex center gap-2.
- **Primary (Charcoal):** `bg-charcoal text-paper-white`; hover is `opacity-90` (no hue shift, no transform). Used on "Sign & Send Transaction", "Deposit".
- **Secondary/Outline:** border + background toggle; no fill by default.
- **Destructive (Signal Red):** `bg-signal-red text-white`; hover `opacity-90`. Used only on "Withdraw".

### Inputs
- **Style:** transparent fill, `border-input` (Fog), `rounded-md`, `h-9`, `shadow-xs`.
- **Focus:** `ring-2 ring-ring/50 ring-offset-2` only — focus announces as a ring, not a border shift. The red ring is reserved for `aria-invalid`/`destructive` states.

### Cards / Containers
- **Shape:** `rounded-md`, white fill, `border-zinc-200`, `shadow-sm`, animated appear (`fade-in slide-in-from-top-4 duration-500`).
- **Padding:** `p-6` header, `pt-6` content.
- **Internal split:** header (title + description), content (form), optional result below.

### Badges
- **Shape:** `rounded-full`, `px-2 py-0.5`, `text-xs` font-medium.
- **Variants:** default (Charcoal), secondary (Silver), destructive (Signal Red).

### Navigation (Header)
- **Surface:** `border-b bg-white`, no shadow.
- **Links:** `hover:text-zinc-700 hover:underline font-bold`; active state darkens to `text-zinc-800` and keeps underline.
- **Connect control:** `px-4 py-2 bg-zinc-900 rounded-md text-zinc-50 text-sm font-semibold`.

## Do's and Don'ts

### Do:
- **Do** keep signal-red for destructive actions only — Withdraw and delete flows, never primary/affirmative buttons.
- **Do** use the single Card pattern (white, `border-zinc-200`, `shadow-sm`, animated appear) for every task panel; consistency is the point in a tooling surface.
- **Do** let focus render as a 2px ring (`ring-ring/50`) on inputs and buttons.

### Don't:
- **Don't** introduce a second accent color; the palette exists to make cryptographic data legible, not to embellish.
- **Don't** drop the 1px card shadow — its absence flattens floating panels against the page field without adding clarity.
- **Don't** use decorative display type or iconography outside `lucide-react`; keep the voice functional.
