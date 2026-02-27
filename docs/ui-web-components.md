# Web Component Library Spec

Package: `@movoz/ui-web` | Source: `packages/ui-web/src/`

## Conventions

- All components use `"use client"` (React client components)
- All components accept a `className` prop for consumer overrides
- Class merging via `cn()` (clsx + tailwind-merge)
- Components with DOM refs use `forwardRef`
- Icons are passed as `ReactNode` props (decoupled from any icon library)
- Theme-aware via CSS variables — components never import from `@movoz/theme`

## Utilities

### `cn(...inputs)`

Merges class names with Tailwind conflict resolution.

```tsx
import { cn } from "@movoz/ui-web";
cn("px-4 py-2", isActive && "bg-zen-subtle", className)
```

---

## Primitives

### Button

Interactive button with variant and size system.

```tsx
import { Button } from "@movoz/ui-web";

<Button variant="primary" size="lg" loading>Submit</Button>
<Button variant="secondary" icon={<Plus />}>Add Item</Button>
<Button variant="ghost" iconRight={<ArrowRight />}>Next</Button>
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `variant` | `"primary" \| "secondary" \| "ghost" \| "danger"` | `"primary"` | Visual style |
| `size` | `"sm" \| "md" \| "lg"` | `"md"` | Size scale |
| `loading` | `boolean` | `false` | Shows spinner, disables button |
| `icon` | `ReactNode` | — | Leading icon |
| `iconRight` | `ReactNode` | — | Trailing icon |

Variant styles:
- **primary**: `bg-zen-text text-zen-bg` — solid, high contrast
- **secondary**: `bg-zen-subtle border border-zen-border` — outlined/subtle
- **ghost**: transparent background, hover reveals subtle bg
- **danger**: `bg-red-500 text-white`

### Text

Polymorphic text component for headings and body text.

```tsx
import { Text } from "@movoz/ui-web";

<Text as="h1" font="serif" weight="bold" size="5xl">Title</Text>
<Text color="muted" size="sm">Subtitle</Text>
<Text as="span" truncate>Overflowing text...</Text>
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `as` | `ElementType` | `"p"` | HTML element to render |
| `size` | `"xs"` through `"5xl"` | `"base"` | Font size |
| `weight` | `"light" \| "normal" \| "medium" \| "semibold" \| "bold"` | `"normal"` | Font weight |
| `color` | `"default" \| "muted" \| "accent"` | `"default"` | Text color |
| `font` | `"sans" \| "serif" \| "ui"` | — | Font family |
| `truncate` | `boolean` | `false` | Truncate with ellipsis |
| `align` | `"left" \| "center" \| "right"` | — | Text alignment |

### Input

Form input with label, error, and helper text.

```tsx
import { Input } from "@movoz/ui-web";

<Input label="Email" type="email" placeholder="you@example.com" />
<Input label="Search" icon={<Search />} size="sm" />
<Input label="Name" error="Required" helperText="Your full name" />
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `label` | `string` | — | Label text (auto-generates `id` for `htmlFor`) |
| `error` | `string` | — | Error message (turns border red) |
| `helperText` | `string` | — | Helper text below input (hidden when error shown) |
| `icon` | `ReactNode` | — | Leading icon inside input |
| `size` | `"sm" \| "md" \| "lg"` | `"md"` | Input size |

### Card

Content container with variant system.

```tsx
import { Card } from "@movoz/ui-web";

<Card>Simple content</Card>
<Card variant="outlined" padding="lg" header={<h3>Title</h3>}>Body</Card>
<Card variant="filled" footer={<Button>Action</Button>}>Content</Card>
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `variant` | `"elevated" \| "outlined" \| "filled"` | `"elevated"` | Visual style |
| `padding` | `"none" \| "sm" \| "md" \| "lg"` | `"md"` | Internal padding |
| `header` | `ReactNode` | — | Header slot (adds border-bottom) |
| `footer` | `ReactNode` | — | Footer slot (adds border-top) |

### Badge

Inline status or label indicator.

```tsx
import { Badge } from "@movoz/ui-web";

<Badge>Default</Badge>
<Badge variant="solid" color="success">Active</Badge>
<Badge variant="outline" color="danger" size="md">Error</Badge>
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `variant` | `"solid" \| "subtle" \| "outline"` | `"subtle"` | Visual style |
| `color` | `"default" \| "accent" \| "success" \| "warning" \| "danger"` | `"default"` | Color theme |
| `size` | `"sm" \| "md"` | `"sm"` | Size |

### Avatar

User image with fallback initials.

```tsx
import { Avatar } from "@movoz/ui-web";

<Avatar src="/photo.jpg" alt="John" />
<Avatar fallback="John Doe" size="lg" />
<Avatar fallback="AB" shape="square" size="xl" />
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `src` | `string` | — | Image URL |
| `alt` | `string` | `""` | Alt text |
| `fallback` | `string` | — | Text for initials (first letter of each word, max 2) |
| `size` | `"sm" \| "md" \| "lg" \| "xl"` | `"md"` | Size (32/40/48/64px) |
| `shape` | `"circle" \| "square"` | `"circle"` | Border shape |

### IconButton

Icon-only button with required accessibility label.

```tsx
import { IconButton } from "@movoz/ui-web";

<IconButton icon={<Menu />} label="Open menu" />
<IconButton icon={<X />} label="Close" variant="ghost" size="sm" />
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `icon` | `ReactNode` | **required** | The icon to render |
| `label` | `string` | **required** | Accessibility label (`aria-label`) |
| `variant` | `"primary" \| "secondary" \| "ghost" \| "danger"` | `"ghost"` | Visual style |
| `size` | `"sm" \| "md" \| "lg"` | `"md"` | Size |

---

## Layout

### Stack

Flexbox layout with direction and gap.

```tsx
import { Stack } from "@movoz/ui-web";

<Stack gap={4}>           {/* vertical, gap-4 */}
<Stack direction="horizontal" gap={2} align="center">
<Stack gap={6} wrap>      {/* wrapping flex */}
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `direction` | `"horizontal" \| "vertical"` | `"vertical"` | Flex direction |
| `gap` | `0-12` | `4` | Gap between items (Tailwind scale) |
| `align` | `"start" \| "center" \| "end" \| "stretch" \| "baseline"` | — | Cross-axis alignment |
| `justify` | `"start" \| "center" \| "end" \| "between" \| "around"` | — | Main-axis alignment |
| `wrap` | `boolean` | `false` | Enable flex wrap |

### Container

Max-width wrapper with centering and padding.

```tsx
import { Container } from "@movoz/ui-web";

<Container>                   {/* max-w-6xl mx-auto px-6 */}
<Container maxWidth="xl">     {/* max-w-7xl */}
<Container padding={false}>   {/* no horizontal padding */}
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `maxWidth` | `"sm" \| "md" \| "lg" \| "xl" \| "2xl" \| "full"` | `"lg"` | Maximum width |
| `padding` | `boolean` | `true` | Horizontal padding (px-6) |
| `centered` | `boolean` | `true` | Center with mx-auto |

Max-width mapping: sm=max-w-2xl, md=max-w-4xl, lg=max-w-6xl, xl=max-w-7xl, 2xl=1400px

### Divider

Horizontal or vertical rule.

```tsx
import { Divider } from "@movoz/ui-web";

<Divider />                        {/* horizontal, default color */}
<Divider color="subtle" />         {/* lighter color */}
<Divider orientation="vertical" /> {/* vertical line */}
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `orientation` | `"horizontal" \| "vertical"` | `"horizontal"` | Direction |
| `color` | `"default" \| "subtle"` | `"default"` | Border color |

### Spacer

Empty space or flex filler.

```tsx
import { Spacer } from "@movoz/ui-web";

<Spacer size={8} />    {/* fixed 2rem space */}
<Spacer flex />        {/* fills remaining space */}
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `size` | `number` | — | Fixed size (Tailwind scale: 4 = 1rem) |
| `flex` | `boolean` | `false` | Fill available space (flex-1) |

---

## Complex

### Modal

Dialog overlay with portal rendering.

```tsx
import { Modal } from "@movoz/ui-web";

const [open, setOpen] = useState(false);

<Modal open={open} onClose={() => setOpen(false)} title="Confirm">
  <p>Are you sure?</p>
</Modal>
```

| Prop | Type | Default | Description |
|---|---|---|---|
| `open` | `boolean` | **required** | Visibility |
| `onClose` | `() => void` | **required** | Close handler |
| `title` | `string` | — | Header title (adds close button) |
| `size` | `"sm" \| "md" \| "lg" \| "xl"` | `"md"` | Width |
| `closeOnOverlay` | `boolean` | `true` | Close on backdrop click |
| `closeOnEsc` | `boolean` | `true` | Close on Escape key |

Features: `createPortal` to `document.body`, backdrop blur, body scroll lock, ESC handler.

### Toast

Notification system with context provider.

```tsx
// Wrap app with provider
<ToastProvider position="bottom-right">
  <App />
</ToastProvider>

// Use in any component
import { useToast } from "@movoz/ui-web";

const { toast, dismiss } = useToast();
toast({ message: "Saved!", variant: "success" });
toast({ message: "Failed", variant: "error", duration: 6000 });
```

**ToastProvider props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `position` | `"top-right" \| "top-left" \| "bottom-right" \| "bottom-left" \| "top-center" \| "bottom-center"` | `"bottom-right"` | Toast position |

**toast() options:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `message` | `string` | **required** | Toast content |
| `variant` | `"success" \| "error" \| "warning" \| "info"` | `"info"` | Color/icon style |
| `duration` | `number` | `4000` | Auto-dismiss time (ms) |

### Dropdown

Trigger-based menu with items.

```tsx
import { Dropdown } from "@movoz/ui-web";

<Dropdown
  trigger={<Button variant="ghost">Options</Button>}
  items={[
    { label: "Edit", value: "edit", icon: <Pencil /> },
    { label: "Delete", value: "delete", destructive: true },
  ]}
  onSelect={(value) => handleAction(value)}
/>
```

**Dropdown props:**

| Prop | Type | Default | Description |
|---|---|---|---|
| `trigger` | `ReactNode` | **required** | Element that opens the menu |
| `items` | `DropdownItem[]` | **required** | Menu items |
| `onSelect` | `(value: string) => void` | **required** | Selection handler |
| `align` | `"start" \| "end"` | `"start"` | Menu alignment |

**DropdownItem:**

| Prop | Type | Description |
|---|---|---|
| `label` | `string` | Display text |
| `value` | `string` | Value passed to `onSelect` |
| `icon` | `ReactNode` | Optional leading icon |
| `disabled` | `boolean` | Disable interaction |
| `destructive` | `boolean` | Red text styling |

Features: click-outside close, ESC close, absolute positioning, no external deps.
