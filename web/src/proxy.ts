/*
 * Волчица и Крылья — middleware (CSP, COMING_SOON).
 * Copyright (C) 2026 Грёзов Саярин Аквилович (https://github.com/ViraMisal)
 *
 * SPDX-License-Identifier: AGPL-3.0-only
 * Лицензия: GNU Affero General Public License v3.0 (только эта версия) —
 * см. https://www.gnu.org/licenses/agpl-3.0.html и файл LICENSE в корне.
 */

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// CSP с per-request nonce: Next видит заголовок CSP в request и сам подставляет
// nonce в свои <script>. Поэтому все страницы рендерятся динамически — в
// закэшированном HTML nonce из запроса оказаться не может.
// Стили — только внешние файлы: ни <style>, ни style="" в JSX нет,
// так что style-src обходится без 'unsafe-inline'.
// COMING_SOON=1: все страницы редиректятся на /soon (переключается в env без пересборки).
export function proxy(req: NextRequest) {
  const nonce = crypto.randomUUID().replaceAll("-", "");
  const isDev = process.env.NODE_ENV !== "production";

  const csp = [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}'${isDev ? " 'unsafe-eval'" : ""}`,
    "style-src 'self'",
    "img-src 'self' data: https:",
    "font-src 'self' data:",
    "connect-src 'self'",
    "object-src 'none'",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join("; ");

  const headers = new Headers(req.headers);
  headers.set("x-nonce", nonce);
  headers.set("Content-Security-Policy", csp);

  if (process.env.COMING_SOON === "1" && req.nextUrl.pathname !== "/soon") {
    const url = req.nextUrl.clone();
    url.pathname = "/soon";
    url.search = "";
    return NextResponse.redirect(url);
  }

  const res = NextResponse.next({ request: { headers } });
  res.headers.set("Content-Security-Policy", csp);
  return res;
}

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|brand|videos).*)",
  ],
};
