A felt-tip outlined button in the Movoz marker font; use for any action. Default is the outlined `secondary`; use `primary` (ink fill) for the main call to action and `accent` (terracotta) sparingly.

```jsx
<Button variant="primary" lift iconRight={<span>→</span>}>Read article</Button>
<Button variant="secondary">Subscribe</Button>
<Button variant="ghost" size="sm">Cancel</Button>
```

Variants: `primary` (ink), `secondary` (outline), `accent` (terracotta), `ghost`. Sizes `sm | md | lg`. Set `lift` for the hard sketch drop-shadow that depresses on press.
