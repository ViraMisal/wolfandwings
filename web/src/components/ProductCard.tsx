"use client";
import Link from "next/link";
import { Brush, Moon, Flame, Archive } from "lucide-react";
import type { Product } from "@/lib/types";
import { formatPrice, STATUS_META } from "@/lib/format";

function StatusBadge({ status }: { status: string }) {
  const m = STATUS_META[status] || STATUS_META.in_stock;
  return (
    <span className={"badge " + m.cls}>
      {m.icon === "dot" ? <span className="dot" /> : null}
      {m.icon === "moon" ? <Moon size={13} /> : null}
      {m.icon === "flame" ? <Flame size={13} /> : null}
      {m.icon === "archive" ? <Archive size={13} /> : null}
      {m.label}
    </span>
  );
}

export function ProductCard({ p }: { p: Product }) {
  const disc = p.old_price ? Math.round((1 - p.base_price / p.old_price) * 100) : 0;

  return (
    <article className={"pcard" + (p.status === "archived" ? " archived" : "")}>
      <Link className="pcard-art" href={"/product/" + p.slug} aria-label={`${p.title} — ${p.artist.name}`}>
        <span className={"art art-tint-" + p.tint}>
          <span className="art-tag">арт · {p.fandom.name}</span>
        </span>
        <StatusBadge status={p.status} />
        {disc ? <span className="badge badge-sale">−{disc}%</span> : null}
      </Link>
      <div className="pcard-body">
        <Link href={"/product/" + p.slug}><span className="pcard-title">{p.title}</span></Link>
        <Link className="pcard-artist" href={"/artists/" + p.artist.slug}>
          <Brush size={14} /> {p.artist.name}
        </Link>
        <span className="pcard-fandom">{p.fandom.name} · {p.category.name}</span>
        <div className="pcard-foot">
          {p.old_price ? (
            <span className="price-row">
              <span className="price-old">{formatPrice(p.old_price)}</span>
              <span className="price">{formatPrice(p.base_price)}</span>
            </span>
          ) : (
            <span className="price">{formatPrice(p.base_price)}</span>
          )}
        </div>
      </div>
    </article>
  );
}
