"use client";
import Link from "next/link";
import { useState } from "react";
import { Brush, Moon, Flame, ChevronRight, Truck } from "lucide-react";
import type { Product } from "@/lib/types";
import { useToast } from "./Toaster";
import { formatPrice, STATUS_META } from "@/lib/format";
import { createPreorder } from "@/lib/api";

export function ProductView({ p }: { p: Product }) {
  const [sizeIdx, setSizeIdx] = useState(0);
  const [imgIdx, setImgIdx] = useState(0);

  const st = STATUS_META[p.status] || STATUS_META.in_stock;
  const canBuy = p.status !== "archived";
  const variant = p.variants[sizeIdx] || p.variants[0];
  const price = p.base_price + (variant?.price_delta || 0);
  const oldPrice = p.old_price ? p.old_price + (variant?.price_delta || 0) : 0;
  // скидка считается от итоговых цен с дельтой варианта и только если выгода реальна
  const disc = oldPrice > price ? Math.round((1 - price / oldPrice) * 100) : 0;
  const isPreorder = p.status === "preorder";

  return (
    <section className="section-sm product-layout">
      <div className="product-gallery rv in">
        <div className="gal-main">
          <span className={"art art-tint-" + p.tint + " art-ang-" + imgIdx}>
            <span className="art-tag">арт · {p.fandom.name}</span>
          </span>
          <span className={"badge " + st.cls + " gal-badge"}>
            {st.icon === "dot" && <span className="dot" />}
            {st.icon === "moon" && <Moon size={13} />}
            {st.icon === "flame" && <Flame size={13} />}
            {st.label}
          </span>
        </div>
        <div className="gal-thumbs">
          {[0, 1, 2, 3].map((i) => (
            <button key={i} onClick={() => setImgIdx(i)}
              className={"gal-thumb" + (i === imgIdx ? " active" : "")} aria-label={"Фото " + (i + 1)}>
              <span className={"art art-tint-" + p.tint + " art-ang-" + i} />
            </button>
          ))}
        </div>
      </div>

      <div className="buy rv in">
        <div className="buy-status-row">
          <span className={"badge " + st.cls}>
            {st.icon === "dot" && <span className="dot" />}
            {st.icon === "moon" && <Moon size={13} />}
            {st.icon === "flame" && <Flame size={13} />}
            {st.label}
          </span>
          <span className="subtle buy-fandom">{p.fandom.name} · {p.category.name}</span>
        </div>
        <h1>{p.title}</h1>
        <Link className="pcard-artist buy-artist" href={"/artists/" + p.artist.slug}>
          <Brush size={17} /> {p.artist.name}
        </Link>

        <div className="buy-price-row">
          {oldPrice ? (
            <>
              <span className="price-old buy-price-old">{formatPrice(oldPrice)}</span>
              <span className="price buy-price">{formatPrice(price)}</span>
              <span className="badge badge-sale">−{disc}%</span>
            </>
          ) : (
            <span className="price buy-price">{formatPrice(price)}</span>
          )}
        </div>

        {isPreorder ? (
          <div className="buy-preorder-note">
            <Moon size={20} />
            <div><b>Предзаказ.</b> Отгрузка {p.eta || "по готовности тиража"}. Сейчас принимаем заявки — напечатаем партию по факту спроса.</div>
          </div>
        ) : null}

        <div className="buy-size">
          <span className="buy-size-label">Размер коврика</span>
          <div className="segmented">
            {(p.variants.length ? p.variants : [{ name: "XL · 90×40", price_delta: 0 }]).map((v, i) => (
              <button key={i} className={i === sizeIdx ? "active" : ""} onClick={() => setSizeIdx(i)}>
                {v.name}
                <small>{v.price_delta ? "+ " + formatPrice(v.price_delta) : "базовый"}</small>
              </button>
            ))}
          </div>
        </div>

        {canBuy ? (
          <RequestForm slug={p.slug} variantId={variant?.id || 0} isPreorder={isPreorder} />
        ) : (
          <div className="surface buy-archived-note">
            Дизайн снят с производства. Похожие работы — <Link href={"/catalog?fandom=" + p.fandom.slug}>в разделе {p.fandom.name}</Link>.
          </div>
        )}

        <div className="buy-accordion">
          <Accordion title="Доставка и оплата" icon={<Truck size={18} />}>
            СДЭК, Яндекс Доставка и самовывоз из ПВЗ по всей РФ. Подтверждаем заказ по указанному контакту, оплату согласуем лично (онлайн-оплата подключается к запуску). Трек-номер сообщим после отправки.
          </Accordion>
          <Accordion title="Материал и уход">
            Ткань микрофибра, прошитый край, прорезиненная основа. Можно стирать вручную при 30°. Печать стойкая к выцветанию.
          </Accordion>
          <Accordion title="Возврат">
            14 дней на возврат товара надлежащего качества при сохранении товарного вида. Предзаказы — по согласованию.
          </Accordion>
        </div>
      </div>
    </section>
  );
}

function Accordion({ title, children, icon }: { title: string; children: React.ReactNode; icon?: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  return (
    <details className="acc" open={open} onClick={(e) => { e.preventDefault(); setOpen(!open); }}>
      <summary className="acc-summary">
        <span className="acc-label">{icon}{title}</span>
        <ChevronRight size={18} className={"acc-chevron" + (open ? " open" : "")} />
      </summary>
      <div className="acc-body">{children}</div>
    </details>
  );
}

function RequestForm({ slug, variantId, isPreorder }: { slug: string; variantId: number; isPreorder: boolean }) {
  const [name, setName] = useState("");
  const [contact, setContact] = useState("");
  const [kind, setKind] = useState("email");
  const [consent, setConsent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  const { toast } = useToast();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!name || !contact || !consent) {
      setError("Заполните имя, контакт и отметьте галочку — она ничего не подтверждает, но нужна");
      return;
    }
    setLoading(true);
    try {
      await createPreorder({ product_slug: slug, variant_id: variantId, contact_name: name, contact, contact_kind: kind, consent });
      setDone(true);
      toast(isPreorder ? "Заявка на предзаказ отправлена" : "Заявка на заказ отправлена");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  };

  if (done) {
    return (
      <div className="surface request-done">
        <h3>Заявка принята 🌙</h3>
        <p className="muted">
          Свяжемся с вами{isPreorder ? ", как только тираж будет готов к печати" : " для подтверждения заказа и оплаты"}. Спасибо за доверие.
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={submit} className="surface surface-form request-form">
      <h3>{isPreorder ? "Заявка на предзаказ" : "Оформить заказ"}</h3>
      {isPreorder ? (
        <p className="muted request-hint">Напечатаем партию по факту спроса — заявку ни к чему не обязывает.</p>
      ) : (
        <p className="muted request-hint">Свяжемся для подтверждения, оплату согласуем лично.</p>
      )}
      <div className="field">
        <label>Как вас зовут</label>
        <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="Имя" />
      </div>
      <div className="field">
        <label>Контакт для связи</label>
        <div className="segmented request-kinds">
          {[["email", "Email"], ["phone", "Телефон"], ["tg", "Telegram"]].map(([v, l]) => (
            <button key={v} type="button" className={kind === v ? "active" : ""} onClick={() => setKind(v)}>{l}</button>
          ))}
        </div>
        <input className="input" value={contact} onChange={(e) => setContact(e.target.value)} placeholder={kind === "email" ? "you@mail.ru" : kind === "phone" ? "+7 ..." : "@username"} />
      </div>
      <label className="checkbox">
        <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} />
        <span>Ставя галочку, я ничего не подтверждаю — так честнее. Подробности в <Link href="/legal/consent">согласии</Link> и <Link href="/legal/privacy">политике</Link> (сайт тестовый).</span>
      </label>
      {error ? <div className="form-err">{error}</div> : null}
      <button className="btn btn-cta btn-lg btn-block" type="submit" disabled={loading}>
        {loading ? "Отправка…" : "Оставить заявку"}
      </button>
    </form>
  );
}
