/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: "/scout",
  assetPrefix: "/scout",
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config"],
  async rewrites() {
    // Dev-only: proxy API calls to the Scout backend so client code can
    // just call same-origin "/api/*" (cookies stay first-party). In
    // production, Nginx performs the equivalent proxy — see the
    // infra/nginx/nginx.conf change in this task.
    const apiOrigin = process.env.SCOUT_API_URL || "http://localhost:8082";
    return [{ source: "/api/:path*", destination: `${apiOrigin}/api/:path*` }];
  },
};

export default nextConfig;
