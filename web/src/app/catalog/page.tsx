import type { Metadata } from "next";
import Link from "next/link";
import { fetchProducts, fetchArtists, fetchFandoms, fetchCategories } from "@/lib/api";
import { CatalogClient } from "@/components/CatalogClient";

// фильтры приходят из URL — рендерим только на сервере по запросу
export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Каталог",
  description: "Каталог авторских ковриков и дескматов. Фильтры по художнику, фандому, категории и наличию.",
  alternates: { canonical: "/catalog" },
};

// белые списки availability/sort — остальное отбрасываем
const AVAILABILITY = ["in_stock", "preorder"] as const;
const SORTS = ["pop", "new", "cheap", "exp"] as const;

export default async function CatalogPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const sp = await searchParams;
  // параметр может быть продублирован (?a=1&a=2) — берём первое значение
  const one = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);

  // справочники нужны и для валидации slug, и для рендера фильтров
  const [{ items: artists }, { items: fandoms }, { items: categories }] = await Promise.all([
    fetchArtists(),
    fetchFandoms(),
    fetchCategories(),
  ]);

  // валидация: неизвестные slug/значения молча игнорируем — это просто вид по умолчанию
  const artist = artists.find((a) => a.slug === one(sp.artist))?.slug;
  const fandom = fandoms.find((f) => f.slug === one(sp.fandom))?.slug;
  const category = categories.find((c) => c.slug === one(sp.category))?.slug;
  const availability = AVAILABILITY.find((v) => v === one(sp.availability));
  const sort = SORTS.find((v) => v === one(sp.sort));

  const params: Record<string, string | number> = { limit: 100 };
  if (artist) params.artist = artist;
  if (fandom) params.fandom = fandom;
  if (category) params.category = category;
  if (availability) params.availability = availability;
  if (sort) params.sort = sort;
  const { items: products } = await fetchProducts(params);

  return (
    <div className="container">
      <nav className="crumbs">
        <Link href="/">Главная</Link>
        <span className="sep">/</span>
        <span>Каталог</span>
      </nav>
      <CatalogClient
        products={products}
        artists={artists}
        fandoms={fandoms}
        categories={categories}
        filters={{
          artist,
          fandom,
          category,
          availability: availability ?? "all",
          sort: sort ?? "pop",
        }}
      />
    </div>
  );
}
