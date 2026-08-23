/* @ds-bundle: {"format":4,"namespace":"MovozDesignSystem_c2bc1d","components":[{"name":"Annotation","sourcePath":"components/core/Annotation.jsx"},{"name":"Avatar","sourcePath":"components/core/Avatar.jsx"},{"name":"Badge","sourcePath":"components/core/Badge.jsx"},{"name":"Button","sourcePath":"components/core/Button.jsx"},{"name":"Card","sourcePath":"components/core/Card.jsx"},{"name":"Input","sourcePath":"components/core/Input.jsx"},{"name":"Logo","sourcePath":"components/core/Logo.jsx"},{"name":"Pill","sourcePath":"components/core/Pill.jsx"},{"name":"PlaceholderBox","sourcePath":"components/core/PlaceholderBox.jsx"},{"name":"Skeleton","sourcePath":"components/core/Skeleton.jsx"},{"name":"Tabs","sourcePath":"components/core/Tabs.jsx"}],"sourceHashes":{"components/core/Annotation.jsx":"dc029c3d8885","components/core/Avatar.jsx":"a6eac037b6c1","components/core/Badge.jsx":"ea0ebede1b68","components/core/Button.jsx":"034ea16e674c","components/core/Card.jsx":"4de203341666","components/core/Input.jsx":"fd5db191b4f2","components/core/Logo.jsx":"0a68837b42da","components/core/Pill.jsx":"7db5d1b035f4","components/core/PlaceholderBox.jsx":"91bf70198369","components/core/Skeleton.jsx":"f214234f7b41","components/core/Tabs.jsx":"36f99ca0fb24","ui_kits/movoz-blog/ArticleScreen.jsx":"16052c531067","ui_kits/movoz-blog/HomeScreen.jsx":"b9067029dd1f","ui_kits/movoz-blog/NavBar.jsx":"55cbc5ee8b99"},"inlinedExternals":[],"unexposedExports":[]} */

(() => {

const __ds_ns = (window.MovozDesignSystem_c2bc1d = window.MovozDesignSystem_c2bc1d || {});

const __ds_scope = {};

(__ds_ns.__errors = __ds_ns.__errors || []);

// components/core/Annotation.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Annotation — the terracotta hand-written margin note from the wireframe
 * ("↳ hero — featured / latest post"). Use it to label regions in mockups,
 * blueprints, and spec sheets.
 */
function Annotation({
  children,
  arrow = true,
  align = 'left',
  style = {},
  ...rest
}) {
  const base = {
    display: 'inline-flex',
    alignItems: 'baseline',
    gap: '7px',
    fontFamily: 'var(--font-marker)',
    fontStyle: 'italic',
    fontWeight: 'var(--fw-medium)',
    fontSize: 'var(--text-md)',
    color: 'var(--terracotta)',
    justifyContent: align === 'right' ? 'flex-end' : 'flex-start',
    ...style
  };
  return /*#__PURE__*/React.createElement("span", _extends({
    style: base
  }, rest), arrow && /*#__PURE__*/React.createElement("span", {
    style: {
      fontStyle: 'normal'
    }
  }, align === 'right' ? '↳' : '↳'), children);
}
Object.assign(__ds_scope, { Annotation });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Annotation.jsx", error: String((e && e.message) || e) }); }

// components/core/Avatar.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Avatar — a circular outlined frame. Shows an image, initials, or an
 * empty pencil outline (the wireframe's bare author bubble).
 */
function Avatar({
  src = null,
  initials = '',
  size = 40,
  style = {},
  ...rest
}) {
  const base = {
    width: `${size}px`,
    height: `${size}px`,
    borderRadius: '50%',
    border: 'var(--border-width) solid var(--line-strong)',
    background: src ? `center/cover no-repeat url(${src})` : 'var(--paper-raised)',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'var(--ink-soft)',
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: `${Math.round(size * 0.4)}px`,
    flex: '0 0 auto',
    overflow: 'hidden',
    ...style
  };
  return /*#__PURE__*/React.createElement("span", _extends({
    style: base
  }, rest), !src && initials);
}
Object.assign(__ds_scope, { Avatar });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Avatar.jsx", error: String((e && e.message) || e) }); }

// components/core/Badge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Badge / Tag — a small status or category marker.
 * `tone` picks the muted semantic palette; `outline` for a hairline-only chip.
 */
function Badge({
  children,
  tone = 'neutral',
  outline = false,
  style = {},
  ...rest
}) {
  const tones = {
    neutral: {
      fg: 'var(--ink-2)',
      bg: 'var(--pencil-light)',
      bd: 'var(--line)'
    },
    accent: {
      fg: 'var(--terracotta-ink)',
      bg: 'var(--terracotta-soft)',
      bd: 'var(--terracotta)'
    },
    success: {
      fg: 'var(--success)',
      bg: 'var(--success-soft)',
      bd: 'var(--success)'
    },
    warning: {
      fg: 'var(--warning)',
      bg: 'var(--warning-soft)',
      bd: 'var(--warning)'
    },
    danger: {
      fg: 'var(--danger)',
      bg: 'var(--danger-soft)',
      bd: 'var(--danger)'
    },
    info: {
      fg: 'var(--info)',
      bg: 'var(--info-soft)',
      bd: 'var(--info)'
    }
  };
  const t = tones[tone] || tones.neutral;
  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '5px',
    fontFamily: 'var(--font-sans)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: 'var(--text-2xs)',
    letterSpacing: 'var(--ls-wide)',
    textTransform: 'uppercase',
    lineHeight: 1,
    padding: '5px 9px',
    borderRadius: 'var(--radius-sm)',
    color: t.fg,
    background: outline ? 'transparent' : t.bg,
    border: `var(--border-hair) solid ${t.bd}`,
    ...style
  };
  return /*#__PURE__*/React.createElement("span", _extends({
    style: base
  }, rest), children);
}
Object.assign(__ds_scope, { Badge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Badge.jsx", error: String((e && e.message) || e) }); }

// components/core/Button.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Button — a felt-tip outlined control.
 * Primary = ink fill, secondary = paper with ink hairline, ghost = bare.
 * The "lift" variant carries the hard sketch drop-shadow.
 */
function Button({
  children,
  variant = 'secondary',
  size = 'md',
  lift = false,
  iconRight = null,
  iconLeft = null,
  disabled = false,
  type = 'button',
  onClick,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: {
      padding: '6px 12px',
      fontSize: 'var(--text-sm)',
      gap: '6px'
    },
    md: {
      padding: '10px 18px',
      fontSize: 'var(--text-md)',
      gap: '8px'
    },
    lg: {
      padding: '14px 26px',
      fontSize: 'var(--text-lg)',
      gap: '10px'
    }
  };
  const variants = {
    primary: {
      background: 'var(--ink)',
      color: 'var(--paper-raised)',
      border: 'var(--border-width) solid var(--ink)'
    },
    secondary: {
      background: 'var(--paper-raised)',
      color: 'var(--ink)',
      border: 'var(--border-width) solid var(--ink)'
    },
    accent: {
      background: 'var(--terracotta)',
      color: '#fff',
      border: 'var(--border-width) solid var(--terracotta)'
    },
    ghost: {
      background: 'transparent',
      color: 'var(--ink)',
      border: 'var(--border-width) solid transparent'
    }
  };
  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: sizes[size].gap,
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: sizes[size].fontSize,
    lineHeight: 1,
    padding: sizes[size].padding,
    borderRadius: 'var(--radius-md)',
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.45 : 1,
    boxShadow: lift ? 'var(--shadow-sketch)' : 'none',
    transition: 'transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out), background var(--dur-fast)',
    transform: 'translate(0,0)',
    ...variants[variant],
    ...style
  };
  const onDown = e => {
    if (!disabled && lift) {
      e.currentTarget.style.transform = 'translate(3px,3px)';
      e.currentTarget.style.boxShadow = 'none';
    }
  };
  const onUp = e => {
    if (!disabled && lift) {
      e.currentTarget.style.transform = 'translate(0,0)';
      e.currentTarget.style.boxShadow = 'var(--shadow-sketch)';
    }
  };
  return /*#__PURE__*/React.createElement("button", _extends({
    type: type,
    disabled: disabled,
    onClick: onClick,
    onMouseDown: onDown,
    onMouseUp: onUp,
    onMouseLeave: onUp,
    style: base
  }, rest), iconLeft, children, iconRight);
}
Object.assign(__ds_scope, { Button });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Button.jsx", error: String((e && e.message) || e) }); }

// components/core/Card.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Card — a paper sheet with a felt-tip hairline border.
 * `lift` adds the hard sketch drop-shadow; `interactive` raises it on hover.
 */
function Card({
  children,
  lift = false,
  interactive = false,
  padding = 'var(--space-5)',
  as = 'div',
  style = {},
  ...rest
}) {
  const Tag = as;
  const [hover, setHover] = React.useState(false);
  const base = {
    background: 'var(--paper-raised)',
    border: 'var(--border-width) solid var(--line)',
    borderRadius: 'var(--radius-md)',
    padding,
    boxShadow: lift ? 'var(--shadow-sketch)' : 'var(--shadow-none)',
    transition: 'transform var(--dur-base) var(--ease-out), box-shadow var(--dur-base) var(--ease-out), border-color var(--dur-base)',
    transform: interactive && hover ? 'translate(-2px,-2px)' : 'translate(0,0)',
    ...(interactive && hover ? {
      boxShadow: 'var(--shadow-sketch)',
      borderColor: 'var(--ink)'
    } : null),
    cursor: interactive ? 'pointer' : 'default',
    ...style
  };
  return /*#__PURE__*/React.createElement(Tag, _extends({
    style: base,
    onMouseEnter: () => interactive && setHover(true),
    onMouseLeave: () => interactive && setHover(false)
  }, rest), children);
}
Object.assign(__ds_scope, { Card });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Card.jsx", error: String((e && e.message) || e) }); }

// components/core/Input.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Input — a felt-tip outlined text field. Optional leading icon (e.g. a
 * search glyph) and pill or default rounding.
 */
function Input({
  value,
  onChange,
  placeholder = '',
  type = 'text',
  iconLeft = null,
  pill = false,
  disabled = false,
  size = 'md',
  style = {},
  wrapStyle = {},
  ...rest
}) {
  const sizes = {
    sm: {
      h: 34,
      fs: 'var(--text-sm)',
      px: 12
    },
    md: {
      h: 42,
      fs: 'var(--text-md)',
      px: 14
    },
    lg: {
      h: 50,
      fs: 'var(--text-lg)',
      px: 16
    }
  };
  const s = sizes[size];
  const wrap = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    height: `${s.h}px`,
    padding: `0 ${s.px}px`,
    background: 'var(--paper-raised)',
    border: 'var(--border-width) solid var(--line)',
    borderRadius: pill ? 'var(--radius-pill)' : 'var(--radius-md)',
    opacity: disabled ? 0.5 : 1,
    transition: 'border-color var(--dur-fast)',
    ...wrapStyle
  };
  const input = {
    flex: 1,
    minWidth: 0,
    border: 'none',
    outline: 'none',
    background: 'transparent',
    color: 'var(--ink)',
    fontFamily: 'var(--font-sans)',
    fontSize: s.fs,
    ...style
  };
  return /*#__PURE__*/React.createElement("label", {
    style: wrap,
    onFocus: e => e.currentTarget.style.borderColor = 'var(--ink)',
    onBlur: e => e.currentTarget.style.borderColor = 'var(--line)'
  }, iconLeft && /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--ink-faint)',
      display: 'inline-flex'
    }
  }, iconLeft), /*#__PURE__*/React.createElement("input", _extends({
    type: type,
    value: value,
    onChange: onChange,
    placeholder: placeholder,
    disabled: disabled,
    style: input
  }, rest)));
}
Object.assign(__ds_scope, { Input });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Input.jsx", error: String((e && e.message) || e) }); }

// components/core/Logo.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Logo — boxed marker initial + wordmark, as drawn in the wireframe
 * header. The mark is an outlined square holding a serif-italic "M".
 */
function Logo({
  wordmark = 'Movoz',
  mark = 'M',
  sublabel = '',
  size = 'md',
  onDark = false,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: {
      box: 28,
      font: 'var(--text-md)',
      word: 'var(--text-lg)'
    },
    md: {
      box: 40,
      font: 'var(--text-xl)',
      word: 'var(--text-2xl)'
    },
    lg: {
      box: 56,
      font: 'var(--text-2xl)',
      word: 'var(--text-3xl)'
    }
  };
  const s = sizes[size];
  const fg = onDark ? 'var(--paper)' : 'var(--ink)';
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: '12px',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: `${s.box}px`,
      height: `${s.box}px`,
      border: `var(--border-width-2) solid ${fg}`,
      borderRadius: 'var(--radius-sm)',
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'var(--font-marker)',
      fontStyle: 'italic',
      fontWeight: 'var(--fw-bold)',
      fontSize: s.font,
      color: fg,
      flex: '0 0 auto'
    }
  }, mark), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      lineHeight: 1
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-marker)',
      fontStyle: 'italic',
      fontWeight: 'var(--fw-bold)',
      fontSize: s.word,
      color: fg
    }
  }, wordmark), sublabel && /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-sans)',
      fontSize: 'var(--text-3xs)',
      letterSpacing: 'var(--ls-label)',
      textTransform: 'uppercase',
      color: onDark ? 'var(--pencil)' : 'var(--ink-faint)',
      marginTop: '4px'
    }
  }, sublabel)));
}
Object.assign(__ds_scope, { Logo });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Logo.jsx", error: String((e && e.message) || e) }); }

// components/core/Pill.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Pill — the rounded filter / chip control from the wireframe's filter row.
 * Active pills get a soft pencil fill; inactive are outlined.
 */
function Pill({
  children,
  active = false,
  as = 'button',
  onClick,
  style = {},
  ...rest
}) {
  const Tag = as;
  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-medium)',
    fontSize: 'var(--text-md)',
    lineHeight: 1,
    padding: '8px 18px',
    borderRadius: 'var(--radius-pill)',
    cursor: 'pointer',
    color: 'var(--ink)',
    background: active ? 'var(--pencil-light)' : 'var(--paper-raised)',
    border: `var(--border-width) solid ${active ? 'var(--line-strong)' : 'var(--line)'}`,
    transition: 'background var(--dur-fast), border-color var(--dur-fast)',
    ...style
  };
  return /*#__PURE__*/React.createElement(Tag, _extends({
    onClick: onClick,
    style: base
  }, rest), children);
}
Object.assign(__ds_scope, { Pill });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Pill.jsx", error: String((e && e.message) || e) }); }

// components/core/PlaceholderBox.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz PlaceholderBox — the crossed "cover image / code sample" frame from the
 * wireframe. A dashed-or-solid rectangle with an X and an optional centered label.
 * Use it anywhere real imagery/content isn't available yet.
 */
function PlaceholderBox({
  label = '',
  ratio = '16 / 9',
  cross = true,
  dashed = false,
  height = null,
  style = {},
  ...rest
}) {
  const wrap = {
    position: 'relative',
    width: '100%',
    aspectRatio: height ? undefined : ratio,
    height: height || undefined,
    background: 'var(--paper-raised)',
    border: `var(--border-width) ${dashed ? 'dashed' : 'solid'} var(--line-strong)`,
    borderRadius: 'var(--radius-sm)',
    overflow: 'hidden',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'var(--ink-faint)',
    fontFamily: 'var(--font-marker)',
    fontStyle: 'italic',
    fontSize: 'var(--text-md)',
    ...style
  };
  const line = {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    color: 'var(--pencil)'
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    style: wrap
  }, rest), cross && /*#__PURE__*/React.createElement("svg", {
    style: line,
    preserveAspectRatio: "none",
    viewBox: "0 0 100 100"
  }, /*#__PURE__*/React.createElement("line", {
    x1: "0",
    y1: "0",
    x2: "100",
    y2: "100",
    stroke: "currentColor",
    strokeWidth: "0.6",
    vectorEffect: "non-scaling-stroke"
  }), /*#__PURE__*/React.createElement("line", {
    x1: "100",
    y1: "0",
    x2: "0",
    y2: "100",
    stroke: "currentColor",
    strokeWidth: "0.6",
    vectorEffect: "non-scaling-stroke"
  })), label && /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'relative',
      background: 'var(--paper-raised)',
      padding: '0 8px'
    }
  }, label));
}
Object.assign(__ds_scope, { PlaceholderBox });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/PlaceholderBox.jsx", error: String((e && e.message) || e) }); }

// components/core/Skeleton.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Skeleton — the soft pencil-gray placeholder bars that stand in for copy.
 * Compose several `Skeleton` lines, or use `lines` for a quick paragraph.
 */
function Skeleton({
  width = '100%',
  height = 12,
  lines = 1,
  gap = 10,
  rounded = true,
  style = {},
  ...rest
}) {
  const bar = (w, key) => /*#__PURE__*/React.createElement("div", {
    key: key,
    style: {
      width: w,
      height: typeof height === 'number' ? `${height}px` : height,
      background: 'var(--pencil)',
      borderRadius: rounded ? 'var(--radius-pill)' : 'var(--radius-xs)'
    }
  });
  if (lines <= 1) {
    return /*#__PURE__*/React.createElement("div", _extends({
      style: {
        ...style
      }
    }, rest), bar(width, 0));
  }

  // Last line is shorter, like real ragged text.
  const widths = Array.from({
    length: lines
  }, (_, i) => i === lines - 1 ? '62%' : typeof width === 'string' ? width : `${width}px`);
  return /*#__PURE__*/React.createElement("div", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: `${gap}px`,
      ...style
    }
  }, rest), widths.map((w, i) => bar(w, i)));
}
Object.assign(__ds_scope, { Skeleton });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Skeleton.jsx", error: String((e && e.message) || e) }); }

// components/core/Tabs.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Movoz Tabs — the segmented control from the wireframe's tool dock.
 * `tone="dock"` renders the dark pill-bar; `tone="light"` for in-page tabs.
 */
function Tabs({
  items = [],
  value,
  onChange,
  tone = 'light',
  style = {},
  ...rest
}) {
  const dark = tone === 'dock';
  const bar = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    padding: '6px',
    borderRadius: 'var(--radius-lg)',
    background: dark ? 'var(--dock)' : 'var(--paper-raised)',
    border: `var(--border-width) solid ${dark ? 'var(--dock-line)' : 'var(--line)'}`,
    ...style
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    style: bar,
    role: "tablist"
  }, rest), items.map(it => {
    const key = typeof it === 'string' ? it : it.value;
    const label = typeof it === 'string' ? it : it.label;
    const accent = typeof it === 'object' && it.accent;
    const active = key === value;
    const tab = {
      fontFamily: 'var(--font-marker)',
      fontWeight: 'var(--fw-semibold)',
      fontSize: 'var(--text-md)',
      lineHeight: 1,
      padding: '8px 16px',
      borderRadius: 'var(--radius-md)',
      cursor: 'pointer',
      border: 'var(--border-hair) solid transparent',
      transition: 'background var(--dur-fast), color var(--dur-fast)',
      background: active ? accent ? 'var(--terracotta)' : dark ? 'var(--paper)' : 'var(--pencil-light)' : 'transparent',
      color: active ? accent ? '#fff' : dark ? 'var(--ink)' : 'var(--ink)' : accent ? 'var(--terracotta)' : dark ? 'var(--pencil)' : 'var(--ink-soft)',
      borderColor: !active && accent ? 'var(--terracotta)' : 'transparent'
    };
    return /*#__PURE__*/React.createElement("button", {
      key: key,
      role: "tab",
      "aria-selected": active,
      onClick: () => onChange && onChange(key),
      style: tab
    }, label);
  }));
}
Object.assign(__ds_scope, { Tabs });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Tabs.jsx", error: String((e && e.message) || e) }); }

// ui_kits/movoz-blog/ArticleScreen.jsx
try { (() => {
// Movoz blog — single article reading view
const {
  Badge,
  Avatar,
  PlaceholderBox,
  Button,
  Card,
  Annotation,
  Skeleton
} = window.MovozDesignSystem_c2bc1d;
function ArticleScreen({
  annotate,
  onBack
}) {
  return /*#__PURE__*/React.createElement("article", {
    style: {
      background: 'var(--paper)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 760,
      margin: '0 auto',
      padding: '40px 24px 80px'
    }
  }, /*#__PURE__*/React.createElement("button", {
    onClick: onBack,
    style: {
      fontFamily: 'var(--font-marker)',
      fontStyle: 'italic',
      fontSize: 16,
      color: 'var(--ink-soft)',
      background: 'none',
      border: 'none',
      cursor: 'pointer',
      padding: 0,
      marginBottom: 22
    }
  }, "\u2190 All articles"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 8,
      marginBottom: 18
    }
  }, /*#__PURE__*/React.createElement(Badge, {
    tone: "accent"
  }, "Featured"), /*#__PURE__*/React.createElement(Badge, null, "Token Security")), /*#__PURE__*/React.createElement("h1", {
    style: {
      fontSize: 46,
      lineHeight: 1.06,
      margin: '0 0 20px'
    }
  }, "How we built token rotation that survives 50M sessions"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 12,
      marginBottom: 28
    }
  }, /*#__PURE__*/React.createElement(Avatar, {
    initials: "MV",
    size: 44
  }), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-marker)',
      fontWeight: 600,
      fontSize: 17
    }
  }, "M. Vance"), /*#__PURE__*/React.createElement("div", {
    className: "movoz-label"
  }, "Jun 22, 2026 \xB7 14 min read"))), annotate && /*#__PURE__*/React.createElement(Annotation, {
    style: {
      marginBottom: 10
    }
  }, "lead figure \u2014 token exchange"), /*#__PURE__*/React.createElement(PlaceholderBox, {
    label: "token exchange diagram",
    ratio: "16 / 9",
    style: {
      marginBottom: 32
    }
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-sans)',
      fontSize: 18,
      lineHeight: 1.6,
      color: 'var(--ink-2)'
    }
  }, /*#__PURE__*/React.createElement("p", null, "Movoz issues short-lived access tokens and rotates refresh tokens on every exchange. When a previously-used refresh token reappears, the whole token family is revoked \u2014 reuse becomes a signal, not a vulnerability."), /*#__PURE__*/React.createElement(Skeleton, {
    lines: 4,
    style: {
      margin: '20px 0'
    }
  }), /*#__PURE__*/React.createElement("h2", {
    style: {
      fontSize: 28,
      margin: '36px 0 14px'
    }
  }, "Rotation under load"), /*#__PURE__*/React.createElement(Skeleton, {
    lines: 5,
    style: {
      margin: '20px 0'
    }
  })), /*#__PURE__*/React.createElement(Card, {
    lift: true,
    style: {
      margin: '34px 0',
      background: 'var(--terracotta-faint)',
      borderColor: 'var(--terracotta)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    className: "movoz-label",
    style: {
      color: 'var(--terracotta-ink)'
    }
  }, "Key takeaway"), /*#__PURE__*/React.createElement("p", {
    style: {
      fontFamily: 'var(--font-marker)',
      fontSize: 22,
      margin: '8px 0 0',
      color: 'var(--ink)'
    }
  }, "Rotate on every exchange. Detect reuse. Revoke the family.")), /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-sans)',
      fontSize: 18,
      lineHeight: 1.6,
      color: 'var(--ink-2)'
    }
  }, /*#__PURE__*/React.createElement(Skeleton, {
    lines: 4
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      marginTop: 44,
      paddingTop: 24,
      borderTop: 'var(--border-width) solid var(--line)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    className: "movoz-label"
  }, "Next \xB7 OAuth 2.1"), /*#__PURE__*/React.createElement(Button, {
    variant: "secondary",
    iconRight: /*#__PURE__*/React.createElement("span", null, "\u2192")
  }, "Anatomy of an access token"))));
}
window.ArticleScreen = ArticleScreen;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/movoz-blog/ArticleScreen.jsx", error: String((e && e.message) || e) }); }

// ui_kits/movoz-blog/HomeScreen.jsx
try { (() => {
// Movoz blog — home / index view (hero + filter + article grid)
const {
  Pill,
  Card,
  Button,
  Badge,
  Avatar,
  PlaceholderBox,
  Annotation,
  Skeleton
} = window.MovozDesignSystem_c2bc1d;
const MOVOZ_POSTS = [{
  cat: 'Architecture',
  min: 9,
  title: 'OIDC discovery at the edge',
  author: 'R. Okafor'
}, {
  cat: 'Token Security',
  min: 12,
  title: 'Refresh-token reuse detection',
  author: 'L. Park'
}, {
  cat: 'OAuth 2.1',
  min: 7,
  title: 'PKCE by default',
  author: 'M. Vance'
}, {
  cat: 'Token Security',
  min: 11,
  title: 'Sender-constrained tokens with DPoP',
  author: 'S. Adeyemi'
}, {
  cat: 'Ways of Working',
  min: 6,
  title: 'A zero-meeting RFC process',
  author: 'J. Wu'
}, {
  cat: 'OAuth 2.1',
  min: 8,
  title: 'Anatomy of an access token',
  author: 'R. Okafor'
}];
function MetaLabel({
  children
}) {
  return /*#__PURE__*/React.createElement("span", {
    className: "movoz-label"
  }, children);
}
function HomeScreen({
  filter,
  setFilter,
  annotate,
  onOpen
}) {
  const filters = ['All', 'OAuth 2.1', 'OpenID Connect', 'Token Security', 'Architecture', 'Ways of Working'];
  const posts = filter === 'All' ? MOVOZ_POSTS : MOVOZ_POSTS.filter(p => p.cat === filter);
  return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("section", {
    style: {
      background: 'var(--paper-sunken)',
      borderBottom: 'var(--border-width) solid var(--line)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1200,
      margin: '0 auto',
      padding: '28px 40px 48px',
      position: 'relative'
    }
  }, annotate && /*#__PURE__*/React.createElement(Annotation, {
    style: {
      marginBottom: 18
    }
  }, "hero \u2014 featured / latest post"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: 56,
      alignItems: 'center',
      marginTop: annotate ? 0 : 14
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement(MetaLabel, null, "Featured \xB7 Token Security \xB7 14 min"), /*#__PURE__*/React.createElement("h1", {
    style: {
      fontSize: 52,
      margin: '16px 0 18px',
      lineHeight: 1.05
    }
  }, "How we built token rotation that survives 50M sessions"), /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 440
    }
  }, /*#__PURE__*/React.createElement(Skeleton, {
    lines: 3
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 12,
      margin: '24px 0 28px'
    }
  }, /*#__PURE__*/React.createElement(Avatar, {
    initials: "MV",
    size: 40
  }), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement(Skeleton, {
    width: 180,
    height: 11
  }))), /*#__PURE__*/React.createElement(Button, {
    variant: "primary",
    lift: true,
    iconRight: /*#__PURE__*/React.createElement("span", null, "\u2192"),
    onClick: () => onOpen && onOpen()
  }, "Read article")), /*#__PURE__*/React.createElement(PlaceholderBox, {
    label: "code sample / token exchange",
    ratio: "5 / 4"
  })))), /*#__PURE__*/React.createElement("section", {
    style: {
      borderBottom: 'var(--border-width) solid var(--line)',
      background: 'var(--paper)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1200,
      margin: '0 auto',
      padding: '18px 40px',
      display: 'flex',
      alignItems: 'center',
      gap: 12,
      flexWrap: 'wrap'
    }
  }, /*#__PURE__*/React.createElement("span", {
    className: "movoz-label",
    style: {
      marginRight: 4
    }
  }, "Filter"), filters.map(f => /*#__PURE__*/React.createElement(Pill, {
    key: f,
    active: filter === f,
    onClick: () => setFilter(f)
  }, f)))), /*#__PURE__*/React.createElement("section", {
    className: "movoz-grid-bg",
    style: {
      background: 'var(--paper)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1200,
      margin: '0 auto',
      padding: '36px 40px 80px'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'baseline',
      justifyContent: 'space-between',
      marginBottom: 22
    }
  }, /*#__PURE__*/React.createElement("h2", {
    style: {
      fontSize: 34,
      margin: 0
    }
  }, "Latest from Movoz"), annotate && /*#__PURE__*/React.createElement(Annotation, {
    align: "right"
  }, "article grid \xB7 3 cols")), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'grid',
      gridTemplateColumns: 'repeat(3, 1fr)',
      gap: 24
    }
  }, posts.map((p, i) => /*#__PURE__*/React.createElement(Card, {
    key: i,
    interactive: true,
    padding: "0",
    onClick: () => onOpen && onOpen()
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      padding: 14
    }
  }, /*#__PURE__*/React.createElement(PlaceholderBox, {
    label: "cover image",
    ratio: "16 / 9"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      padding: '4px 18px 20px'
    }
  }, /*#__PURE__*/React.createElement(MetaLabel, null, p.cat, " \xB7 ", p.min, " min"), /*#__PURE__*/React.createElement("h3", {
    style: {
      fontSize: 22,
      margin: '10px 0 12px'
    }
  }, p.title), /*#__PURE__*/React.createElement(Skeleton, {
    lines: 2
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 10,
      marginTop: 16
    }
  }, /*#__PURE__*/React.createElement(Avatar, {
    size: 28
  }), /*#__PURE__*/React.createElement(Skeleton, {
    width: 120,
    height: 10
  })))))))));
}
window.HomeScreen = HomeScreen;
window.MetaLabel = MetaLabel;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/movoz-blog/HomeScreen.jsx", error: String((e && e.message) || e) }); }

// ui_kits/movoz-blog/NavBar.jsx
try { (() => {
// Movoz blog — top navigation bar
const {
  Logo,
  Input,
  Button
} = window.MovozDesignSystem_c2bc1d;
function SearchIcon() {
  return /*#__PURE__*/React.createElement("svg", {
    width: "16",
    height: "16",
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: "2",
    strokeLinecap: "round"
  }, /*#__PURE__*/React.createElement("circle", {
    cx: "11",
    cy: "11",
    r: "7"
  }), /*#__PURE__*/React.createElement("line", {
    x1: "21",
    y1: "21",
    x2: "16.5",
    y2: "16.5"
  }));
}
function NavBar({
  onNav,
  active = 'Articles'
}) {
  const links = ['Articles', 'Topics', 'Open Source', 'Careers'];
  return /*#__PURE__*/React.createElement("header", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 24,
      padding: '18px 40px',
      borderBottom: 'var(--border-width) solid var(--line)',
      background: 'var(--paper)',
      position: 'sticky',
      top: 0,
      zIndex: 20
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      cursor: 'pointer'
    },
    onClick: () => onNav && onNav('Home')
  }, /*#__PURE__*/React.createElement(Logo, {
    sublabel: "logo + wordmark"
  })), /*#__PURE__*/React.createElement("nav", {
    style: {
      display: 'flex',
      gap: 26,
      marginLeft: 28
    }
  }, links.map(l => /*#__PURE__*/React.createElement("a", {
    key: l,
    onClick: () => onNav && onNav(l),
    style: {
      fontFamily: 'var(--font-marker)',
      fontStyle: 'italic',
      fontSize: 'var(--text-lg)',
      fontWeight: active === l ? 700 : 500,
      color: active === l ? 'var(--ink)' : 'var(--ink-soft)',
      textDecoration: 'none',
      cursor: 'pointer'
    }
  }, l))), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: 'auto',
      display: 'flex',
      gap: 12,
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: 260
    }
  }, /*#__PURE__*/React.createElement(Input, {
    pill: true,
    placeholder: "Search\u2026",
    iconLeft: /*#__PURE__*/React.createElement(SearchIcon, null)
  })), /*#__PURE__*/React.createElement(Button, {
    variant: "secondary"
  }, "Subscribe")));
}
window.NavBar = NavBar;
window.SearchIcon = SearchIcon;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/movoz-blog/NavBar.jsx", error: String((e && e.message) || e) }); }

__ds_ns.Annotation = __ds_scope.Annotation;

__ds_ns.Avatar = __ds_scope.Avatar;

__ds_ns.Badge = __ds_scope.Badge;

__ds_ns.Button = __ds_scope.Button;

__ds_ns.Card = __ds_scope.Card;

__ds_ns.Input = __ds_scope.Input;

__ds_ns.Logo = __ds_scope.Logo;

__ds_ns.Pill = __ds_scope.Pill;

__ds_ns.PlaceholderBox = __ds_scope.PlaceholderBox;

__ds_ns.Skeleton = __ds_scope.Skeleton;

__ds_ns.Tabs = __ds_scope.Tabs;

})();
