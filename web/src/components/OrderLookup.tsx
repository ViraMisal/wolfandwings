"use client";
import Link from "next/link";
import { useState } from "react";
import { Search, Package } from "lucide-react";
import { formatPrice } from "@/lib/format";

interface OrderEvent { status: string; note: string; created_at: string }
interface OrderItem { title: string; artist: string; variant_name: string; price: number; qty: number; line_total: number }
interface Order {
  number: string; status: string; payment_status: string; is_preorder: boolean;
  customer_email: string; total: number; eta: string; delivery_method: string;
  tracking_number: string; created_at: string; items: OrderItem[]; events: OrderEvent[];
}

const STATUS_LABEL: Record<string, string> = {
  created: "Создан",
  awaiting_payment: "Ожидает оплаты",
  paid: "Оплачен",
  in_production: "В производстве",
  packing: "Комплектуется",
  shipped: "Передан в доставку",
  delivered: "Доставлен",
  closed: "Закрыт",
  cancelled: "Отменён",
  refunded: "Возврат",
};

export function OrderLookup() {
  const [number, setNumber] = useState("");
  const [email, setEmail] = useState("");
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const lookup = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setOrder(null);
    if (!number || !email) { setError("Укажите номер заказа и email"); return; }
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/orders/${encodeURIComponent(number)}?email=${encodeURIComponent(email)}`);
      if (!res.ok) {
        const b = await res.json().catch(() => ({}));
        throw new Error((b as { error?: string }).error || "Заказ не найден");
      }
      setOrder(await res.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <nav className="crumbs"><Link href="/">Главная</Link><span className="sep">/</span><span>Статус заказа</span></nav>
      <section className="section-sm lookup-wrap">
        <div className="shead"><div className="stitle"><span className="label-caps">Отслеживание</span><h2>Статус заказа</h2></div></div>
        <form onSubmit={lookup} className="surface surface-form">
          <div className="field">
            <label>Номер заказа</label>
            <input className="input" value={number} onChange={(e) => setNumber(e.target.value)} placeholder="WW-2026-0001" />
          </div>
          <div className="field">
            <label>Email, указанный при оформлении</label>
            <input className="input" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@mail.ru" />
          </div>
          {error ? <div className="form-err">{error}</div> : null}
          <button className="btn btn-luna btn-lg btn-block" type="submit" disabled={loading}>
            {loading ? "Поиск…" : <>Найти заказ <Search size={18} /></>}
          </button>
        </form>

        {order ? (
          <div className="surface lookup-result">
            <div className="between lookup-head">
              <h3 className="mono">{order.number}</h3>
              <span className="badge badge-info">{STATUS_LABEL[order.status] || order.status}</span>
            </div>
            <p className="subtle mono lookup-date">от {order.created_at}</p>

            {order.events.length ? (
              <div className="timeline lookup-timeline">
                {order.events.map((e, i) => (
                  <div key={i} className={"tl-step " + (i === order.events.length - 1 ? "current" : "done")}>
                    <span className="tl-line" />
                    <span className="tl-dot"><Package size={16} /></span>
                    <span className="tl-label">{STATUS_LABEL[e.status] || e.status}</span>
                    <span className="tl-date">{e.created_at.slice(0, 10)}</span>
                  </div>
                ))}
              </div>
            ) : null}

            <div className="divider divider-lg" />
            {order.items.map((it, i) => (
              <div key={i} className="between lookup-item">
                <span>{it.title} <span className="subtle lookup-item-meta">· {it.artist} · {it.variant_name} ×{it.qty}</span></span>
                <span className="mono">{formatPrice(it.line_total)}</span>
              </div>
            ))}
            <div className="divider divider-sm" />
            <div className="between"><span className="muted">Итого</span><span className="price">{formatPrice(order.total)}</span></div>
            {order.tracking_number ? <p className="mono subtle lookup-track">Трек: {order.tracking_number}</p> : null}
          </div>
        ) : null}
      </section>
    </>
  );
}
