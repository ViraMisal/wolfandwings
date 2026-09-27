import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchProduct, fetchProducts } from "@/lib/api";
import { ProductView } from "@/components/ProductView";
import { ProductCard } from "@/components/ProductCard";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  try {
    const p = await fetchProduct(slug);
    return {
      title: p.title,
      description: `${p.title} — авторский коврик от ${p.artist.name}. ${p.fandom.name}.`,
      alternates: { canonical: "/product/" + slug },
      openGraph: { title: p.title, description: `${p.artist.name} · ${p.fandom.name}` },
    };
  } catch {
    return { title: "Товар не найден" };
  }
}

export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  let p;
  try {
    p = await fetchProduct(slug);
  } catch {
    notFound();
  }

  const more = (await fetchProducts({ artist: p.artist.slug, limit: 100 })).items.filter((x) => x.slug !== slug).slice(0, 4);

  return (
    <div className="container">
      <nav className="crumbs">
        <Link href="/">Главная</Link>
        <span className="sep">/</span>
        <Link href="/catalog">Каталог</Link>
        <span className="sep">/</span>
        <a href={"/catalog?fandom=" + p.fandom.slug}>{p.fandom.name}</a>
        <span className="sep">/</span>
        <span>{p.title}</span>
      </nav>
      <ProductView p={p} />
      {more.length ? (
        <section className="section-sm">
          <div className="shead">
            <div className="stitle"><span className="label-caps">Тот же автор</span><h2>Ещё от {p.artist.name}</h2></div>
            <a className="btn btn-ghost" href={"/artists/" + p.artist.slug}>Профиль автора →</a>
          </div>
          <div className="cards-grid">{more.map((x) => <ProductCard key={x.slug} p={x} />)}</div>
        </section>
      ) : null}
    </div>
  );
}
