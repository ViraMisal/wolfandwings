import type {
  ArtistFull,
  PageDoc,
  Product,
  ProductsResponse,
  Ref,
} from "./types";

const API_INTERNAL = process.env.API_INTERNAL_URL || "http://127.0.0.1:8080";
const PUBLIC_API = process.env.NEXT_PUBLIC_API_URL || "/api";

/** Абсолютный базовый URL: на сервере — прямой к Go, на клиенте — относительный /api (через rewrite). */
function base(): string {
  return typeof window === "undefined" ? API_INTERNAL + "/api" : PUBLIC_API;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(base() + path, {
    headers: { Accept: "application/json" },
    // на сервере достаточно свежих данных (минута кеша), на клиенте — всегда точные
    next: typeof window === "undefined" ? { revalidate: 60 } : { revalidate: 0 },
  });
  if (!res.ok) throw new Error(`API ${res.status}: ${path}`);
  return res.json() as Promise<T>;
}

export async function fetchProducts(params: Record<string, string | number> = {}): Promise<ProductsResponse> {
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) qs.set(k, String(v));
  const q = qs.toString();
  return get<ProductsResponse>("/v1/catalog/products" + (q ? "?" + q : ""));
}

export async function fetchProduct(slug: string): Promise<Product> {
  return get<Product>("/v1/catalog/products/" + encodeURIComponent(slug));
}

export async function fetchArtists(): Promise<{ items: ArtistFull[] }> {
  return get("/v1/catalog/artists");
}

export async function fetchArtist(slug: string): Promise<ArtistFull> {
  return get<ArtistFull>("/v1/catalog/artists/" + encodeURIComponent(slug));
}

export async function fetchFandoms(): Promise<{ items: Ref[] }> {
  return get("/v1/catalog/fandoms");
}

export async function fetchCategories(): Promise<{ items: Ref[] }> {
  return get("/v1/catalog/categories");
}

export async function fetchPage(slug: string): Promise<PageDoc> {
  return get<PageDoc>("/v1/catalog/pages/" + encodeURIComponent(slug));
}

/** Создание заявки-предзаказа (клиентский POST). */
export async function createPreorder(body: {
  product_slug: string;
  variant_id: number;
  contact_name: string;
  contact: string;
  contact_kind: string;
  consent: boolean;
}): Promise<{ id: number; status: string }> {
  const res = await fetch(PUBLIC_API + "/v1/preorders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || "Не удалось отправить заявку");
  }
  return res.json();
}
