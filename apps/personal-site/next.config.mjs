/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config", "@movoz/ui-web"],
  async rewrites() {
    return {
      beforeFiles: [
        { source: "/daily", destination: "http://localhost:3001/daily" },
        { source: "/daily/:path*", destination: "http://localhost:3001/daily/:path*" },
      ],
    };
  },
};

export default nextConfig;
