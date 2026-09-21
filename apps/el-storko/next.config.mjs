const apiUrl = process.env.EL_STORKO_API_URL || "http://localhost:8082";

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config", "@movoz/ui-web"],
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiUrl}/api/:path*` }];
  },
};

export default nextConfig;
