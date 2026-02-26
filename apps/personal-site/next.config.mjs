/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config", "@movoz/ui-web"],
};

export default nextConfig;
