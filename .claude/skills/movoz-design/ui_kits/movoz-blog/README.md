# Movoz Blog — UI Kit

A high-fidelity recreation of the Movoz engineering/marketing blog in the lo-fi
"annotated wireframe" house style. This is the canonical product surface the
design system was reverse-engineered from.

## Files
- `index.html` — interactive demo. Switch **Home ⇄ Article** and toggle the
  **Annotations** and **Grid** overlays from the floating tool dock (bottom).
- `NavBar.jsx` — sticky header: boxed logo lockup, italic marker nav, pill search, Subscribe.
- `HomeScreen.jsx` — hero (featured post + code-sample placeholder), filter pill
  row, 3-column article grid. Filter pills actually filter the list.
- `ArticleScreen.jsx` — single-article reader: badges, author, lead figure,
  body copy with skeleton blocks, terracotta key-takeaway callout.

## Composition
Screens compose only design-system primitives from `window.MovozDesignSystem_*`
(`Logo`, `Input`, `Button`, `Pill`, `Card`, `Badge`, `Avatar`, `PlaceholderBox`,
`Annotation`, `Skeleton`, `Tabs`). No primitive is re-implemented here.

## Notes / intentionally lo-fi
Body copy is represented with `Skeleton` bars exactly as in the source wireframe;
imagery uses `PlaceholderBox`. Swap these for real content when productionizing.
