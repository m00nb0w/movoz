/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config"],
};

export default nextConfig;
