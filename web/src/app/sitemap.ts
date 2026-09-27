import type { MetadataRoute } from "next";
import { fetchProducts, fetchArtists } from "@/lib/api";

const SITE = process.env.NEXT_PUBLIC_SITE_URL || "https://wolfandwings.ru";

// динамически: при статической сборке API ещё нет и sitemap ушёл бы без товаров
export const dynamic = "force-dynamic";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base: MetadataRoute.Sitemap = [
    { url: SITE + "/", changeFrequency: "weekly", priority: 1 },
    { url: SITE + "/catalog", changeFrequency: "daily", priority: 0.9 },
    { url: SITE + "/artists", changeFrequency: "weekly", priority: 0.7 },
    { url: SITE + "/legal/oferta", changeFrequency: "yearly", priority: 0.2 },
    { url: SITE + "/legal/privacy", changeFrequency: "yearly", priority: 0.2 },
    { url: SITE + "/legal/consent", changeFrequency: "yearly", priority: 0.2 },
    { url: SITE + "/legal/requisites", changeFrequency: "yearly", priority: 0.2 },
    { url: SITE + "/legal/returns", changeFrequency: "yearly", priority: 0.2 },
    { url: SITE + "/legal/about", changeFrequency: "yearly", priority: 0.3 },
  ];

  try {
    const [{ items: products }, { items: artists }] = await Promise.all([
      fetchProducts({ limit: 100 }),
      fetchArtists(),
    ]);
    for (const p of products) base.push({ url: `${SITE}/product/${p.slug}`, changeFrequency: "weekly", priority: 0.8 });
    for (const a of artists) base.push({ url: `${SITE}/artists/${a.slug}`, changeFrequency: "weekly", priority: 0.6 });
  } catch {}

  return base;
}
