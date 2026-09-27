import type { Metadata } from "next";
import Link from "next/link";
import { fetchArtists } from "@/lib/api";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Художники",
  description: "Витрина приглашённых художников «Волчица и Крылья».",
  alternates: { canonical: "/artists" },
};

export default async function ArtistsPage() {
  const { items: artists } = await fetchArtists();
  return (
    <div className="container">
      <nav className="crumbs"><Link href="/">Главная</Link><span className="sep">/</span><span>Художники</span></nav>
      <div className="section-sm pt-0">
        <div className="shead"><div className="stitle"><span className="label-caps">Витрина авторов</span><h2>Наши художники</h2></div></div>
        <div className="artists-list">
          {artists.map((a) => (
            <Link key={a.slug} className="artist-card" href={"/artists/" + a.slug}>
              <span className="avatar ring art avatar-64 art-tint-luna" />
              <span>
                <span className="a-name">{a.name}</span>
                <br />
                <span className="a-meta">{a.bio}</span>
              </span>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
