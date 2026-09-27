import Link from "next/link";
import { Truck, RefreshCw } from "lucide-react";
import { fetchProducts, fetchArtists } from "@/lib/api";
import { ProductCard } from "@/components/ProductCard";
import { Reveal } from "@/components/Reveal";

// SSR на сервере (где живёт API), не SSG при сборке — сборка не зависит от API.
export const dynamic = "force-dynamic";

const FANDOMS = [
  { name: "Arknights", slug: "arknights", tint: "luna", meta: "Резкое серебро" },
  { name: "ZZZ", slug: "zzz", tint: "amber", meta: "Неон и драйв" },
  { name: "HSR", slug: "hsr", tint: "violet", meta: "Звёздный путь" },
  { name: "Genshin", slug: "genshin", tint: "teal", meta: "Стихии и свет" },
];

export default async function HomePage() {
  const [{ items: products }, { items: artists }] = await Promise.all([
    fetchProducts({ limit: 8 }),
    fetchArtists(),
  ]);
  const featured = [
    ...products.filter((p) => p.status === "preorder"),
    ...products.filter((p) => p.status !== "preorder"),
  ].slice(0, 4);

  return (
    <>
      <section className="hero container">
        <div className="hero-grid">
          <Reveal>
            <span className="eyebrow">Авторский арт · лунная ночь</span>
            <h1 className="display hero-title">
              <span className="display-line">Носи ночь</span>
              <span className="display-line-luna">на своём столе</span>
            </h1>
            <p className="muted hero-lead">
              Коврики и дескматы с оригинальным фан-артом по любимым гача-играм. Каждый дизайн — работа приглашённого художника.
            </p>
            <div className="btn-row">
              <Link className="btn btn-cta btn-lg" href="/catalog">Смотреть коврики</Link>
              <Link className="btn btn-outline btn-lg" href="/artists">Наши художники</Link>
            </div>
            <div className="hero-stats">
              <div className="stat">
                <b className="stat-num">4</b>
                <span className="stat-cap">фандома</span>
              </div>
              <div className="stat">
                <b className="stat-num">{products.length}+</b>
                <span className="stat-cap">авторских работ</span>
              </div>
              <div className="stat">
                <b className="stat-num">СДЭК</b>
                <span className="stat-cap">Яндекс Доставка · ПВЗ</span>
              </div>
            </div>
          </Reveal>
          <div className="rv in hero-art" aria-hidden="true">
            <div className="hero-art-glow" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="hero-art-img" src="/brand/mark.svg" alt="" />
          </div>
        </div>
      </section>

      <section className="section-sm container">
        <div className="shead">
          <div className="stitle"><span className="label-caps">Свежее серебро</span><h2>Новинки и предзаказы</h2></div>
          <Link className="btn btn-ghost" href="/catalog">Весь каталог →</Link>
        </div>
        <div className="cards-grid">
          {featured.map((p) => <ProductCard key={p.slug} p={p} />)}
        </div>
      </section>

      <section className="section-sm container">
        <div className="shead"><div className="stitle"><span className="label-caps">По вселенным</span><h2>Выбирай свой фандом</h2></div></div>
        <div className="fandom-grid">
          {FANDOMS.map((f) => (
            <Link key={f.slug} className="cat-card" href={"/catalog?fandom=" + f.slug}>
              <span className={"art art-fill art-tint-" + f.tint} />
              <span className="cat-card-shade" />
              <span className="cat-card-text">
                <span className="c-name">{f.name}</span>
                <br />
                <span className="cat-card-meta">{f.meta}</span>
              </span>
            </Link>
          ))}
        </div>
      </section>

      <section className="section-sm container">
        <div className="shead">
          <div className="stitle"><span className="label-caps">Витрина авторов</span><h2>Художники</h2></div>
          <Link className="btn btn-ghost" href="/artists">Все авторы →</Link>
        </div>
        <div className="artists-grid">
          {artists.map((a) => (
            <Link key={a.slug} className="artist-card" href={"/artists/" + a.slug}>
              <span className="avatar ring art avatar-60 art-tint-luna" />
              <span>
                <span className="a-name">{a.name}</span>
                <br />
                <span className="a-meta">{a.bio.split(".")[0]}</span>
              </span>
            </Link>
          ))}
        </div>
      </section>

      <section className="section-sm container">
        <div className="trust">
          <div className="t-item"><Truck /><div><h4>Доставка по РФ</h4><p>СДЭК, Яндекс Доставка и самовывоз из ПВЗ. Трек-номер — к сообщению об отправке.</p></div></div>
          <div className="t-item"><RefreshCw /><div><h4>Возврат по закону</h4><p>Дистанционная покупка: 7 дней на отказ, деньги — в течение 10.</p></div></div>
        </div>
      </section>
    </>
  );
}
