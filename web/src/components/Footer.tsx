import Link from "next/link";

export function Footer() {
  return (
    <footer className="site-footer">
      <div className="container">
        <div className="footer-grid">
          <div>
            <Link href="/" className="brand footer-brand">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img className="logo-mark" src="/brand/mark.svg" alt="" width={34} height={34} />
              <span className="wm">
                <b>Волчица</b> <span>и&nbsp;Крылья</span>
              </span>
            </Link>
            <p className="muted footer-about">
              Магазин авторского фан-арта. Коврики и дескматы от приглашённых художников.
            </p>
          </div>
          <div>
            <h4>Магазин</h4>
            <Link href="/catalog">Каталог</Link>
            <Link href="/catalog?availability=preorder">Предзаказы</Link>
            <Link href="/artists">Художники</Link>
            <Link href="/order">Статус заказа</Link>
          </div>
          <div>
            <h4>Помощь</h4>
            <Link href="/legal/returns">Доставка и возврат</Link>
            <Link href="/legal/about">О магазине</Link>
            <Link href="/legal/requisites">Реквизиты</Link>
          </div>
          <div>
            <h4>Документы</h4>
            <Link href="/legal/oferta">Оферта</Link>
            <Link href="/legal/privacy">Политика ПД</Link>
            <Link href="/legal/consent">Согласие на обработку</Link>
            <Link href="/legal/requisites">Реквизиты</Link>
          </div>
        </div>
        <div className="footer-bottom">
          <span>© 2026 Волчица и&nbsp;Крылья · Реквизиты — на странице «Реквизиты»</span>
          <Link href="/roll" className="mono" aria-label="Наши контакты">wolfandwings.ru</Link>
        </div>
      </div>
    </footer>
  );
}
