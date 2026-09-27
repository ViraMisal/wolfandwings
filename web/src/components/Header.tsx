"use client";
import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { Menu, X } from "lucide-react";

const NAV: { label: string; href: string }[] = [
  { label: "Каталог", href: "/catalog" },
  { label: "Художники", href: "/artists" },
  { label: "Предзаказы", href: "/catalog?availability=preorder" },
  { label: "Статус заказа", href: "/order" },
];

function NavLinks() {
  const pathname = usePathname();
  const search = useSearchParams()?.toString();

  // ссылка с параметрами активна только при точном совпадении запроса,
  // чистая — только когда в адресе нет чужих параметров
  const isCurrent = (href: string) => {
    const [path, qs] = href.split("?");
    if (pathname !== path) return false;
    return qs ? search === qs : !search;
  };

  return (
    <nav className="main-nav">
      {NAV.map((n) => (
        <Link key={n.href} href={n.href} aria-current={isCurrent(n.href) ? "page" : undefined}>
          {n.label}
        </Link>
      ))}
    </nav>
  );
}

export function Header() {
  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <>
      <header className={"site-header" + (scrolled ? " scrolled" : "")}>
        <div className="header-in">
          <Link href="/" className="brand" aria-label="Волчица и Крылья — на главную">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="logo-mark" src="/brand/mark.svg" alt="" width={36} height={36} />
            <span className="wm">
              <b>Волчица</b> <span>и&nbsp;Крылья</span>
            </span>
          </Link>
          <Suspense fallback={<nav className="main-nav" />}>
            <NavLinks />
          </Suspense>
          <div className="header-tools">
            <button className="icon-btn burger" aria-label="Меню" onClick={() => setMenuOpen(true)}>
              <Menu size={22} />
            </button>
          </div>
        </div>
      </header>

      <div className={"mobile-menu" + (menuOpen ? " open" : "")}>
        <div className="mm-top">
          <Link href="/" className="brand" onClick={() => setMenuOpen(false)}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="logo-mark" src="/brand/mark.svg" alt="" width={32} height={32} />
            <span className="wm">
              <b>Волчица</b> <span>и&nbsp;Крылья</span>
            </span>
          </Link>
          <button className="icon-btn" aria-label="Закрыть" onClick={() => setMenuOpen(false)}>
            <X size={22} />
          </button>
        </div>
        <nav>
          {NAV.map((n) => (
            <Link key={n.href} href={n.href} onClick={() => setMenuOpen(false)}>
              {n.label}
            </Link>
          ))}
        </nav>
      </div>
    </>
  );
}
