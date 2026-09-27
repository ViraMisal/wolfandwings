import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchArtist } from "@/lib/api";
import { plural } from "@/lib/format";
import { ProductCard } from "@/components/ProductCard";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  try {
    const a = await fetchArtist(slug);
    return { title: a.name, description: a.bio, alternates: { canonical: "/artists/" + slug } };
  } catch {
    return { title: "Художник не найден" };
  }
}

export default async function ArtistPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  let a;
  try {
    a = await fetchArtist(slug);
  } catch {
    notFound();
  }
  return (
    <div className="container">
      <nav className="crumbs"><Link href="/">Главная</Link><span className="sep">/</span><Link href="/artists">Художники</Link><span className="sep">/</span><span>{a.name}</span></nav>
      <section className="section-sm pt-0">
        <div className="surface artist-head">
          <span className="avatar ring art avatar-96" />
          <div>
            <h1>{a.name}</h1>
            <p className="muted artist-bio">{a.bio}</p>
            <p className="subtle mono artist-count">{a.works_count} работ</p>
          </div>
        </div>
        <div className="shead"><div className="stitle"><span className="label-caps">Портфолио</span><h2>{a.works_count} {plural(a.works_count, ["работа", "работы", "работ"])}</h2></div></div>
        <div className="cards-grid">
          {a.products.map((p) => <ProductCard key={p.slug} p={p} />)}
        </div>
      </section>
    </div>
  );
}
