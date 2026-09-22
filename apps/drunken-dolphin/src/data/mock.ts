// Typed mock data ported from wiki/design/drunken-dolphin/data.jsx
// Used by Slice 1 (Dashboard). Replace with real API responses in Slice 5.

export type Bucket = "essentials" | "lifestyle" | "irregular";
export type Reaction = "loved" | "boring" | "kill" | null;

export type DigestItem = {
  title: string;
  source: string;
  minutes: number;
  tag: string | null;
  summary: string;
  reaction: Reaction;
};

export type DigestTopic = {
  name: string;
  slug: string;
  count: number;
  items: DigestItem[];
};

export type DigestToday = {
  date: string;
  label: string;
  summary: string;
  itemCount: number;
  unread: number;
  topics: DigestTopic[];
};

export type DigestArchiveEntry = {
  date: string;
  items: number;
  summary: string;
  topics: string[];
};

export type Expense = {
  id?: number;
  date?: string;
  time: string;
  text: string;
  amount: number;
  bucket: Bucket;
  category: string;
  tag: string | null;
};

export type BucketSummary = {
  bucket: Bucket;
  label: string;
  amount: number;
  share: number;
  categories: string[];
};

export type CategorySummary = {
  name: string;
  amount: number;
  bucket: Bucket;
};

export type TagRollup = {
  name: string;
  total: number;
  count: number;
  period: string;
};

export const TODAY = {
  weekday: "Sunday",
  date: "May 03, 2026",
  short: "Sun May 03",
};

export const DIGEST_TODAY: DigestToday = {
  date: "2026-05-03",
  label: "Sunday, May 3 · 7:00 AM",
  summary: "Quiet weekend in markets. Two notable auth posts. SBV kept rates unchanged.",
  itemCount: 14,
  unread: 9,
  topics: [
    {
      name: "Auth & Identity",
      slug: "auth",
      count: 3,
      items: [
        {
          title: "Why FIDO2 attestation matters more than you think",
          source: "security.googleblog",
          minutes: 6,
          tag: "P1",
          summary:
            "A long-overdue look at why platform attestation is finally reaching maturity, with a real-world bypass case study from a fintech.",
          reaction: null,
        },
        {
          title: "Apple deprecates SCEP, pushes ACME for managed devices",
          source: "developer.apple.com",
          minutes: 3,
          tag: null,
          summary:
            "Apple is moving the entire managed-device certificate flow to ACME, leaving SCEP on a deprecation path through 2027.",
          reaction: "loved",
        },
        {
          title: "A formal model of OAuth 2.1 in TLA+",
          source: "arxiv.org",
          minutes: 12,
          tag: "long",
          summary:
            "A reader-submitted formal verification of the new OAuth 2.1 draft. Finds two underspecified edge cases around resource indicators.",
          reaction: null,
        },
      ],
    },
    {
      name: "AI",
      slug: "ai",
      count: 4,
      items: [
        {
          title: "Anthropic releases Sonnet 4.5 with agent SDK",
          source: "anthropic.com",
          minutes: 4,
          tag: null,
          summary:
            "New model and a first-party agent SDK that exposes tool-use and computer-use as a single primitive.",
          reaction: null,
        },
        {
          title: 'A skeptical reading of "AI 2027"',
          source: "lesswrong",
          minutes: 18,
          tag: "long",
          summary:
            "A long but careful response to last week's viral forecasting essay. Disagrees in mostly the right places.",
          reaction: null,
        },
        {
          title: "Local LLM throughput on M5 Max — 2× the M3 at half the watts",
          source: "simonwillison.net",
          minutes: 5,
          tag: null,
          summary:
            "Benchmark dump from a week of testing. Llama 3.3 70B at 38 tok/s on a laptop is the headline.",
          reaction: "boring",
        },
        {
          title: "OpenAI files for IPO at $850B valuation",
          source: "reuters.com",
          minutes: 2,
          tag: null,
          summary:
            "Reuters confirms the F-1. Operating margin negative. Cost of revenue mostly compute.",
          reaction: null,
        },
      ],
    },
    {
      name: "Tech",
      slug: "tech",
      count: 2,
      items: [
        {
          title: "PostgreSQL 18 ships with native asynchronous replication slots",
          source: "postgresql.org",
          minutes: 4,
          tag: null,
          summary:
            "The headline ergonomic feature in PG18. Removes the most common reason for using third-party CDC tools.",
          reaction: null,
        },
        {
          title: "Why your CDN is the wrong cache",
          source: "fly.io/blog",
          minutes: 9,
          tag: null,
          summary:
            "Walks through why a CDN sits in the wrong place for most application caching, and what to put there instead.",
          reaction: null,
        },
      ],
    },
    {
      name: "Finance & Markets",
      slug: "finance",
      count: 3,
      items: [
        {
          title: "SBV holds policy rate at 4.5%",
          source: "cafef.vn",
          minutes: 2,
          tag: "VN",
          summary:
            "State Bank of Vietnam holds rates for the third consecutive meeting, citing dong stability.",
          reaction: null,
        },
        {
          title: "Why everyone is suddenly talking about non-bank lending",
          source: "matt-levine",
          minutes: 11,
          tag: null,
          summary:
            "Matt Levine on the migration of credit risk out of the banking system, and what regulators are missing.",
          reaction: null,
        },
        {
          title: "Q1 earnings: tech beats by 4%, retail misses by 2%",
          source: "wsj.com",
          minutes: 3,
          tag: null,
          summary:
            "A clean split. Cloud and ads led the upside; consumer-discretionary names mostly disappointed.",
          reaction: null,
        },
      ],
    },
    {
      name: "Vietnam",
      slug: "vn",
      count: 2,
      items: [
        {
          title: "HCMC metro line 2 groundbreaking pushed to Q3",
          source: "tuoitre.vn",
          minutes: 2,
          tag: "VN",
          summary:
            "Original translation: city committee cites land-clearing for the third delay this year.",
          reaction: null,
        },
        {
          title: "Foreign tech hires in Vietnam up 31% YoY",
          source: "vneconomy",
          minutes: 4,
          tag: "VN",
          summary:
            "Vietnam's tech sector pulled in record foreign hires in Q1, with Da Nang outpacing HCMC for the first time.",
          reaction: null,
        },
      ],
    },
  ],
};

export const DIGEST_ARCHIVE: DigestArchiveEntry[] = [
  { date: "Sat May 02", items: 11, summary: "Quiet day. Sonnet 4.5 leaks. SBV signal piece.", topics: ["ai", "finance"] },
  { date: "Fri May 01", items: 16, summary: "Big day for AI infra. Two long auth reads.", topics: ["ai", "auth", "tech"] },
  { date: "Thu Apr 30", items: 13, summary: "Q1 earnings preview. New OIDC draft.", topics: ["finance", "auth"] },
  { date: "Wed Apr 29", items: 9, summary: "Light news cycle. Local LLM benchmarks.", topics: ["ai"] },
  { date: "Tue Apr 28", items: 14, summary: "Vietnam GDP print. Postgres 18 RC.", topics: ["vn", "tech"] },
  { date: "Mon Apr 27", items: 12, summary: "Hacker News best of the week roundup.", topics: ["tech", "ai"] },
  { date: "Sun Apr 26", items: 10, summary: "Sunday read: long Matt Levine.", topics: ["finance"] },
  { date: "Sat Apr 25", items: 8, summary: "Slow Saturday. One auth ICYMI.", topics: ["auth"] },
];

export const TODAY_EXPENSES: Expense[] = [
  { id: 1, time: "08:42", text: "Cà phê sữa", amount: 45_000, bucket: "lifestyle", category: "Coffee", tag: null },
  { id: 2, time: "12:30", text: "Lunch — bún chả", amount: 120_000, bucket: "essentials", category: "Food", tag: null },
  { id: 3, time: "14:15", text: "Grab to D2", amount: 38_000, bucket: "essentials", category: "Transport", tag: null },
  { id: 4, time: "19:20", text: "Bookstore — TLA+ book", amount: 380_000, bucket: "lifestyle", category: "Books", tag: null },
];

export const MONTH_BY_BUCKET: BucketSummary[] = [
  { bucket: "essentials", label: "Essentials", amount: 8_420_000, share: 0.42, categories: ["Food", "Transport", "Rent", "Utilities"] },
  { bucket: "lifestyle", label: "Lifestyle", amount: 6_850_000, share: 0.34, categories: ["Coffee", "Books", "Eating out", "Hobbies"] },
  { bucket: "irregular", label: "Irregular", amount: 4_780_000, share: 0.24, categories: ["Trip", "Gifts", "Health"] },
];

export const MONTH_BUDGET = 22_000_000;
export const MONTH_SPENT = 20_050_000;
export const MONTH_DAYS_IN = 3;
export const MONTH_DAYS_TOTAL = 31;

export const DAILY_SPEND = [
  410, 290, 880, 1240, 540, 720,
  680, 1100, 583,
];

export const WEEKLY_TREND = [
  4_200_000, 5_100_000, 4_800_000, 6_300_000, 5_900_000, 4_500_000,
  5_700_000, 6_200_000, 5_400_000, 4_900_000, 6_100_000, 5_200_000,
];

export const CATEGORIES: CategorySummary[] = [
  { name: "Food", amount: 4_120_000, bucket: "essentials" },
  { name: "Rent", amount: 3_500_000, bucket: "essentials" },
  { name: "Transport", amount: 620_000, bucket: "essentials" },
  { name: "Utilities", amount: 180_000, bucket: "essentials" },
  { name: "Coffee", amount: 1_240_000, bucket: "lifestyle" },
  { name: "Books", amount: 980_000, bucket: "lifestyle" },
  { name: "Eating out", amount: 3_400_000, bucket: "lifestyle" },
  { name: "Hobbies", amount: 1_230_000, bucket: "lifestyle" },
  { name: "Trip · #ustravel2026", amount: 3_200_000, bucket: "irregular" },
  { name: "Gifts", amount: 780_000, bucket: "irregular" },
  { name: "Health", amount: 800_000, bucket: "irregular" },
];

export const TAGS: TagRollup[] = [
  { name: "#ustravel2026", total: 18_400_000, count: 22, period: "since Feb" },
  { name: "#wedding", total: 6_200_000, count: 8, period: "Mar–Apr" },
  { name: "#camera", total: 4_800_000, count: 5, period: "Apr" },
];

export const RECENT_EXPENSES: Expense[] = [
  { date: "May 03", time: "19:20", text: "Bookstore — TLA+ book", amount: 380_000, bucket: "lifestyle", category: "Books", tag: null },
  { date: "May 03", time: "14:15", text: "Grab to D2", amount: 38_000, bucket: "essentials", category: "Transport", tag: null },
  { date: "May 03", time: "12:30", text: "Lunch — bún chả", amount: 120_000, bucket: "essentials", category: "Food", tag: null },
  { date: "May 03", time: "08:42", text: "Cà phê sữa", amount: 45_000, bucket: "lifestyle", category: "Coffee", tag: null },
  { date: "May 02", time: "21:05", text: "Dinner — sushi w/ Mai", amount: 680_000, bucket: "lifestyle", category: "Eating out", tag: null },
  { date: "May 02", time: "16:30", text: "Visa fee — US tourist", amount: 4_100_000, bucket: "irregular", category: "Trip", tag: "#ustravel2026" },
  { date: "May 02", time: "11:00", text: "Pharmacy", amount: 95_000, bucket: "irregular", category: "Health", tag: null },
  { date: "May 01", time: "20:00", text: "Movie + popcorn", amount: 220_000, bucket: "lifestyle", category: "Hobbies", tag: null },
  { date: "May 01", time: "13:20", text: "Pho", amount: 80_000, bucket: "essentials", category: "Food", tag: null },
  { date: "May 01", time: "09:00", text: "Electricity bill", amount: 920_000, bucket: "essentials", category: "Utilities", tag: null },
  { date: "Apr 30", time: "22:10", text: "Cocktail bar — birthday", amount: 540_000, bucket: "lifestyle", category: "Eating out", tag: null },
  { date: "Apr 30", time: "12:00", text: "Lunch — com tam", amount: 75_000, bucket: "essentials", category: "Food", tag: null },
  { date: "Apr 30", time: "08:30", text: "Gym monthly", amount: 700_000, bucket: "lifestyle", category: "Hobbies", tag: null },
  { date: "Apr 29", time: "19:40", text: "Domain renewal", amount: 320_000, bucket: "irregular", category: "Hobbies", tag: null },
];
