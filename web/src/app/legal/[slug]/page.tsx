import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { fetchPage } from "@/lib/api";
import { renderMarkdown } from "@/lib/markdown";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  try {
    const p = await fetchPage(slug);
    return { title: p.title, description: p.meta_description || p.title, alternates: { canonical: "/legal/" + slug } };
  } catch {
    return { title: "Документ" };
  }
}

export default async function LegalPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  let p;
  try {
    p = await fetchPage(slug);
  } catch {
    notFound();
  }
  return (
    <div className="container">
      <nav className="crumbs">
        <Link href="/">Главная</Link><span className="sep">/</span><span>{p.title}</span>
      </nav>
      <section className="section-sm pt-0">
        <div className="prose-legal" dangerouslySetInnerHTML={{ __html: renderMarkdown(p.body) }} />
      </section>
    </div>
  );
}
