import type { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement>;

const base = (size = 14): SVGProps<SVGSVGElement> => ({
  width: size,
  height: size,
  viewBox: "0 0 16 16",
  fill: "none",
});

export const Icon = {
  search: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <circle cx="7" cy="7" r="4.5" stroke="currentColor" strokeWidth="1.4" />
      <path d="M11 11l3 3" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  ),
  plus: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <path d="M8 3v10M3 8h10" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  ),
  arrowUp: (p: IconProps = {}) => (
    <svg {...base(11)} {...p}>
      <path d="M5 9l3-3 3 3" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  arrowDown: (p: IconProps = {}) => (
    <svg {...base(11)} {...p}>
      <path d="M5 7l3 3 3-3" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  arrowRight: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <path d="M3 8h10M9 4l4 4-4 4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  external: (p: IconProps = {}) => (
    <svg {...base(11)} {...p}>
      <path d="M6 3h7v7M13 3L7 9M11 13H4a1 1 0 0 1-1-1V5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  heart: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <path d="M8 13.5S2 10 2 5.8C2 4.3 3.2 3 4.7 3c1 0 1.9.5 2.5 1.3l.8 1 .8-1C9.4 3.5 10.3 3 11.3 3 12.8 3 14 4.3 14 5.8 14 10 8 13.5 8 13.5z" stroke="currentColor" strokeWidth="1.3" />
    </svg>
  ),
  yawn: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <circle cx="8" cy="8" r="6" stroke="currentColor" strokeWidth="1.3" />
      <path d="M5.5 6.5h1M9.5 6.5h1M6 10c.5.5 1.2.8 2 .8s1.5-.3 2-.8" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
    </svg>
  ),
  kill: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <path d="M3 5h10M6 5V3.5A.5.5 0 0 1 6.5 3h3a.5.5 0 0 1 .5.5V5M5 5l.7 7.5a1 1 0 0 0 1 .9h2.6a1 1 0 0 0 1-.9L11 5" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  bookmark: (p: IconProps = {}) => (
    <svg {...base()} {...p}>
      <path d="M4 3h8v11l-4-2.5L4 14V3z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round" />
    </svg>
  ),
};
