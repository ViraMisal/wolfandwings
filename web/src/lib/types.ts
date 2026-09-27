export interface ArtistRef {
  slug: string;
  name: string;
}
export interface Ref {
  slug: string;
  name: string;
}
export interface Variant {
  id: number;
  name: string;
  sku: string;
  price_delta: number;
  stock_qty: number;
}
export interface ImageDTO {
  url: string;
  alt: string;
  width: number;
  height: number;
  is_primary: boolean;
}
export type ProductStatus = "in_stock" | "preorder" | "low" | "archived" | "draft";
export interface Product {
  slug: string;
  title: string;
  description: string;
  base_price: number;
  old_price?: number;
  currency: string;
  status: ProductStatus;
  eta?: string;
  tint: string;
  artist: ArtistRef;
  fandom: Ref;
  category: Ref;
  images: ImageDTO[];
  variants: Variant[];
  featured?: boolean;
}
export interface ProductsResponse {
  items: Product[];
  count: number;
  filters: Record<string, string>;
}
export interface ArtistFull {
  slug: string;
  name: string;
  bio: string;
  avatar_url: string;
  socials: Record<string, string>;
  works_count: number;
  products: Product[];
}
export interface PageDoc {
  slug: string;
  title: string;
  body: string;
  meta_description: string;
}