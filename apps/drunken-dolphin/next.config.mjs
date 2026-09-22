/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: "/daily",
  assetPrefix: "/daily",
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config", "@movoz/ui-web"],
};

export default nextConfig;
