/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: "/scout",
  assetPrefix: "/scout",
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config"],
  async rewrites() {
    // Dev-only: proxy API calls to the Scout backend so client code can
    // just call same-origin "/api/*" (cookies stay first-party). In
    // production, Nginx will perform the equivalent proxy — that route is not
    // live yet: Scout deployment (Dockerfiles + compose services + nginx
    // routes) is a deferred follow-up plan, see the TODO block at the top of
    // infra/nginx/nginx.conf and wiki/specs/scout.md's Open Questions.
    const apiOrigin = process.env.SCOUT_API_URL || "http://localhost:8082";
    return [{ source: "/api/:path*", destination: `${apiOrigin}/api/:path*` }];
  },
};

export default nextConfig;
