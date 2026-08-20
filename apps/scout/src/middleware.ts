import { NextRequest, NextResponse } from "next/server";

// Scout's "no public route group" rule (NF1) applies to the frontend too:
// every page redirects to /login unless a session cookie is present. This
// is a UX redirect only — the backend's RequireAuth middleware is the
// authoritative gate (this middleware only checks cookie *presence*, not
// signature validity, since it has no access to SESSION_SECRET).
const SESSION_COOKIE = "scout_session";

export function middleware(request: NextRequest) {
  const isLoginPage = request.nextUrl.pathname === "/login";
  const hasSession = request.cookies.has(SESSION_COOKIE);

  if (!hasSession && !isLoginPage) {
    const loginUrl = request.nextUrl.clone();
    loginUrl.pathname = "/login";
    return NextResponse.redirect(loginUrl);
  }
  if (hasSession && isLoginPage) {
    const homeUrl = request.nextUrl.clone();
    homeUrl.pathname = "/";
    return NextResponse.redirect(homeUrl);
  }
  return NextResponse.next();
}

export const config = {
  // Two entries, not one, because of a Next.js basePath quirk: the
  // basePath ("/scout") is prepended to each matcher source as a raw
  // string, so "/((?!...).*)" becomes "/scout/((?!...).*)" — which
  // requires a "/" right after "/scout" and therefore never matches the
  // bare zone root request "/scout" itself (verified empirically against
  // the dev server). Next specifically special-cases a literal "/" entry
  // to make the trailing segment optional, so it's listed separately here
  // to make sure the root page is actually gated.
  //
  // "api" is excluded because /api/* requests are proxied straight through
  // to the Go backend (via the dev rewrite / prod Nginx route) and are
  // gated there by RequireAuth, not here. Without this exclusion, an
  // unauthenticated POST /api/auth/login (i.e. login itself) would get
  // redirected to /login before ever reaching the backend, making login
  // impossible.
  matcher: ["/", "/((?!_next/static|_next/image|favicon.ico|api).*)"],
};
