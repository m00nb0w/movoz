import type { Config } from "tailwindcss";
import sharedConfig from "@movoz/tailwind-config";

const config: Partial<Config> = {
  presets: [sharedConfig as Config],
};

export default config;
