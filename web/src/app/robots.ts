import type { MetadataRoute } from "next";

const SITE = process.env.NEXT_PUBLIC_SITE_URL || "https://wolfandwings.ru";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: { userAgent: "*", allow: "/", disallow: ["/api/", "/order"] },
    sitemap: SITE + "/sitemap.xml",
    host: "wolfandwings.ru",
  };
}
