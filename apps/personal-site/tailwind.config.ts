import type { Config } from "tailwindcss";
import sharedConfig from "@movoz/tailwind-config";

const config: Config = {
  presets: [sharedConfig as Config],
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/theme/src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/ui-web/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  plugins: [],
};

export default config;
