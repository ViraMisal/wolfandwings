import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Скоро откроемся",
  description: "«Волчица и Крылья» — магазин авторского фан-арта на ковриках. Сайт в разработке.",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

export default function SoonPage() {
  return (
    <section className="container soon-hero">
      <div className="soon-in">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img className="soon-mark" src="/brand/mark.svg" alt="" width={96} height={96} />
        <span className="eyebrow">Сайт в разработке</span>
        <h1 className="display soon-title">
          <span className="display-line">Скоро</span>
          <span className="display-line-luna">откроемся</span>
        </h1>
        <p className="muted soon-lead">
          «Волчица и Крылья» — авторские коврики и дескматы с фан-артом по любимым гача-играм.
          Каждый дизайн — работа приглашённого художника.
        </p>
        <div className="btn-row-center">
          <a className="btn btn-luna btn-lg" href="/roll">Следить в Telegram</a>
          <a className="btn btn-outline btn-lg" href="/roll">Мы во ВКонтакте</a>
        </div>
        <p className="subtle mono soon-note">wolfandwings.ru · скоро</p>
      </div>
    </section>
  );
}
