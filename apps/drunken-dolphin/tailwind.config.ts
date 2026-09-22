import type { Config } from "tailwindcss";
import sharedConfig from "@movoz/tailwind-config";

const config: Config = {
  presets: [sharedConfig as Config],
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/theme/src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/ui-web/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        bucket: {
          essentials: "var(--bucket-essentials)",
          lifestyle: "var(--bucket-lifestyle)",
          irregular: "var(--bucket-irregular)",
        },
      },
    },
  },
  plugins: [],
};

export default config;
