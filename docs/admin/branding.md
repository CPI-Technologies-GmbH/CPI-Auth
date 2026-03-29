# Branding

The Branding page controls the visual appearance of all user-facing authentication pages (login, signup, password reset, etc.). Changes apply to the entire tenant and are reflected in real-time through the live preview.

## Color Pickers

Four color pickers control the core color palette of your authentication pages.

### Primary Color

The main accent color used for buttons, links, and interactive elements. Displayed as the primary call-to-action color throughout the authentication flow.

- Default: `#4F46E5` (indigo)
- Used for: submit buttons, active form elements, links
- CSS property: `--af-color-primary`

### Secondary Color

A complementary accent color for secondary UI elements, hover states, and decorative accents.

- Default: `#7C3AED` (violet)
- Used for: secondary buttons, highlights, badges
- CSS property: `--af-color-secondary`

### Background Color

The page background color for the authentication screens.

- Default: `#FFFFFF` (white)
- Used for: page background, card backgrounds
- CSS property: `--af-color-background`

### Text Color

The primary text color used throughout the authentication pages.

- Default: `#1F2937` (dark gray)
- Used for: headings, body text, labels
- CSS property: `--af-color-text`

Each color picker supports:

- Click to open a visual color picker
- Direct hex input field
- HSL and RGB value inputs
- Recently used colors palette

---

## Logo Upload

Upload separate logos for light and dark backgrounds to ensure visibility in all contexts.

### Light Mode Logo

Displayed on light backgrounds. Typically a dark or colored version of your logo.

- Supported formats: PNG, SVG, JPG
- Recommended size: 200x60px or similar horizontal ratio
- Maximum file size: 2MB
- CSS property: `--af-logo-url`

### Dark Mode Logo

Displayed on dark backgrounds. Typically a white or light-colored version of your logo.

- Same format and size requirements as light mode
- Falls back to light mode logo if not set
- CSS property: `--af-logo-dark-url`

Upload area supports drag-and-drop or click-to-browse. A preview of the uploaded logo is shown inline.

---

## Font Family Selector

Choose the font used across all authentication pages. The selector provides a dropdown of available font families.

| Font | Style |
|------|-------|
| Inter | Modern sans-serif (default) |
| System UI | OS native font stack |
| Roboto | Google's clean sans-serif |
| Open Sans | Friendly, readable sans-serif |
| Lato | Balanced, professional sans-serif |
| Poppins | Geometric, modern sans-serif |
| Source Sans Pro | Adobe's versatile sans-serif |
| Nunito | Rounded, approachable sans-serif |
| Montserrat | Elegant, geometric headings |
| Raleway | Thin, stylish display font |

The font preview updates in real-time as you change the selection. Custom fonts can be loaded through page template CSS.

CSS property: `--af-font-family`

---

## Border Radius Slider

Controls the roundness of buttons, inputs, cards, and other UI elements.

- **Range**: 0px (sharp corners) to 24px (very rounded)
- **Default**: 8px
- **Slider** with numeric input for precise values
- CSS property: `--af-border-radius`

Preview:

| Value | Appearance |
|-------|------------|
| 0px | Square corners, sharp geometric look |
| 4px | Subtle rounding, professional feel |
| 8px | Moderate rounding (default), balanced |
| 12px | Noticeably rounded, friendly feel |
| 16px | Very rounded, soft appearance |
| 24px | Pill-shaped buttons and inputs |

---

## Layout Mode

Select the overall layout structure for authentication pages.

### Centered

The default layout. The login/signup form is centered horizontally and vertically on the page with the logo above it.

```
┌────────────────────────────────┐
│                                │
│           [  Logo  ]           │
│         ┌──────────┐           │
│         │   Form   │           │
│         │          │           │
│         └──────────┘           │
│                                │
└────────────────────────────────┘
```

### Split-Screen

The page is divided into two halves. The left side shows a branded image or gradient, and the right side contains the form.

```
┌───────────────┬────────────────┐
│               │                │
│   Branded     │   [  Logo  ]   │
│   Image or    │  ┌──────────┐  │
│   Gradient    │  │   Form   │  │
│               │  │          │  │
│               │  └──────────┘  │
│               │                │
└───────────────┴────────────────┘
```

### Sidebar

The form appears in a fixed sidebar on the left, with the main area showing branding content.

```
┌──────────┬─────────────────────┐
│          │                     │
│ [ Logo ] │                     │
│┌────────┐│   Branded Content   │
││  Form  ││                     │
││        ││                     │
│└────────┘│                     │
│          │                     │
└──────────┴─────────────────────┘
```

---

## Live Preview

The right side of the branding page displays a live preview that updates instantly as you modify any setting. The preview shows:

- The login page rendered with current branding settings
- Accurate font rendering
- Logo placement according to the selected layout mode
- Button and input styling with the configured border radius
- Color scheme applied to all elements

### Preview Controls

| Control | Description |
|---------|-------------|
| **Page selector** | Switch between login, signup, password reset previews |
| **Dark mode toggle** | Preview how pages look with dark backgrounds |
| **Viewport size** | Toggle between desktop and mobile preview widths |

---

## Saving Changes

Changes are not applied automatically. Click the **"Save Changes"** button at the bottom of the page to persist your branding settings. A confirmation toast appears on success.

To revert unsaved changes, click **"Reset"** to return to the last saved state.

---

## Generated CSS Properties

The branding configuration generates the following CSS custom properties, available in all page templates:

```css
:root {
  --af-color-primary: #4F46E5;
  --af-color-secondary: #7C3AED;
  --af-color-background: #FFFFFF;
  --af-color-text: #1F2937;
  --af-font-family: 'Inter', sans-serif;
  --af-border-radius: 8px;
  --af-logo-url: url('https://cdn.myapp.com/logo.png');
  --af-logo-dark-url: url('https://cdn.myapp.com/logo-dark.png');
}
```

These properties are used by all default templates and should be used in custom templates for consistent styling. See the [Design Tokens](/cli/design-tokens) documentation for the full token reference.
