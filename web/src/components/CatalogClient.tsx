"use client";
import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { SlidersHorizontal, X, Search } from "lucide-react";
import type { ArtistFull, Product, Ref } from "@/lib/types";
import { ProductCard } from "./ProductCard";
import { plural } from "@/lib/format";

type Availability = "all" | "in_stock" | "preorder";
type Sort = "pop" | "new" | "cheap" | "exp";

/** Текущие фильтры — источник истины URL, приходят с сервера в props. */
export type CatalogFilters = {
  artist?: string;
  fandom?: string;
  category?: string;
  availability: Availability;
  sort: Sort;
};

const SORT_LABELS: [Sort, string][] = [
  ["pop", "по популярности"],
  ["new", "сначала новинки"],
  ["cheap", "сначала дешевле"],
  ["exp", "сначала дороже"],
];

export function CatalogClient({
  products,
  artists,
  fandoms,
  categories,
  filters,
}: {
  products: Product[];
  artists: ArtistFull[];
  fandoms: Ref[];
  categories: Ref[];
  filters: CatalogFilters;
}) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [sheetOpen, setSheetOpen] = useState(false);

  const artistList = artists.map((a) => ({ slug: a.slug, name: a.name }));

  // применить фильтры: пересобрать URL /catalog?... и заменить его без скролла
  const apply = (next: Partial<CatalogFilters>) => {
    const f = { ...filters, ...next };
    const params = new URLSearchParams();
    if (f.artist) params.set("artist", f.artist);
    if (f.fandom) params.set("fandom", f.fandom);
    if (f.category) params.set("category", f.category);
    if (f.availability !== "all") params.set("availability", f.availability);
    if (f.sort !== "pop") params.set("sort", f.sort); // pop — значение по умолчанию, в URL не пишем
    const qs = params.toString();
    const cur = toQuery(filters);
    if (qs === cur) return; // ничего не изменилось — навигация не нужна
    startTransition(() => {
      router.replace("/catalog" + (qs ? `?${qs}` : ""), { scroll: false });
    });
  };

  // одиночный выбор: повторный клик по выбранному пункту снимает его
  const pick = (key: "artist" | "fandom" | "category", slug: string) =>
    apply({ [key]: filters[key] === slug ? undefined : slug });

  const chips: [string, string][] = [];
  if (filters.artist) chips.push(["artist", artistList.find((a) => a.slug === filters.artist)?.name || filters.artist]);
  if (filters.fandom) chips.push(["fandom", fandoms.find((f) => f.slug === filters.fandom)?.name || filters.fandom]);
  if (filters.category) chips.push(["category", categories.find((c) => c.slug === filters.category)?.name || filters.category]);
  if (filters.availability !== "all") chips.push(["availability", filters.availability === "preorder" ? "Предзаказ" : "В наличии"]);

  const resetAll = () => apply({ artist: undefined, fandom: undefined, category: undefined, availability: "all" });
  const removeChip = (k: string) => {
    if (k === "availability") apply({ availability: "all" });
    else apply({ [k]: undefined } as Partial<CatalogFilters>);
  };

  // scope — префикс name у radio-групп: сайдбар и мобильный шит в DOM одновременно
  const renderFilters = (scope: string) => (
    <>
      <FilterGroup title="Художник">
        {artistList.map((a) => (
          <Opt key={a.slug} name={`${scope}-artist`} label={a.name} checked={filters.artist === a.slug} onSelect={() => pick("artist", a.slug)} />
        ))}
      </FilterGroup>
      <FilterGroup title="Фандом">
        {fandoms.map((f) => (
          <Opt key={f.slug} name={`${scope}-fandom`} label={f.name} checked={filters.fandom === f.slug} onSelect={() => pick("fandom", f.slug)} />
        ))}
      </FilterGroup>
      <FilterGroup title="Категория">
        {categories.map((c) => (
          <Opt key={c.slug} name={`${scope}-category`} label={c.name} checked={filters.category === c.slug} onSelect={() => pick("category", c.slug)} />
        ))}
      </FilterGroup>
      <FilterGroup title="Наличие">
        {([["all", "Все"], ["in_stock", "В наличии"], ["preorder", "Предзаказ"]] as [Availability, string][]).map(([v, l]) => (
          <Opt key={v} name={`${scope}-availability`} label={l} checked={filters.availability === v} onSelect={() => apply({ availability: v })} />
        ))}
      </FilterGroup>
    </>
  );

  const title = filters.fandom
    ? fandoms.find((f) => f.slug === filters.fandom)?.name || filters.fandom
    : filters.availability === "preorder"
      ? "Предзаказы"
      : filters.availability === "in_stock"
        ? "В наличии"
        : "Все коврики";

  return (
    <>
      <div className="between catalog-head">
        <div className="stitle">
          <h1>{title}</h1>
          <span className="mono subtle catalog-count">
            {products.length} {plural(products.length, ["товар", "товара", "товаров"])}
          </span>
        </div>
        <div className="catalog-tools">
          <button className="btn btn-outline btn-sm filters-btn" onClick={() => setSheetOpen(true)}>
            <SlidersHorizontal size={16} /> фильтры
          </button>
          <label className="subtle catalog-count" htmlFor="sortsel">Сортировка</label>
          <select className="select catalog-sort" id="sortsel" value={filters.sort} onChange={(e) => apply({ sort: e.target.value as Sort })}>
            {SORT_LABELS.map(([v, l]) => (
              <option key={v} value={v}>{l}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="section-sm catalog-layout">
        <aside className="filters-desktop">{renderFilters("d")}</aside>
        <div>
          <div className="catalog-chips">
            {chips.map(([k, label]) => (
              <button key={k} className="chip active" onClick={() => removeChip(k)}>
                {label} <X size={14} />
              </button>
            ))}
            {chips.length ? <button className="chip" onClick={resetAll}>Сбросить всё</button> : null}
          </div>
          {products.length ? (
            <div className={"cards-grid catalog-grid" + (pending ? " catalog-grid-pending" : "")}>
              {products.map((p) => <ProductCard key={p.slug} p={p} />)}
            </div>
          ) : (
            <div className="empty">
              <Search className="e-ico" />
              <p>Ничего не нашлось под выбранные фильтры.</p>
              <button className="btn btn-outline" onClick={resetAll}>Сбросить фильтры</button>
            </div>
          )}
        </div>
      </div>

      {/* мобильный bottom sheet: фильтры применяются сразу к URL, кнопка закрывает шит */}
      <div className={"overlay" + (sheetOpen ? " open" : "")} onClick={() => setSheetOpen(false)} />
      <div className={"drawer drawer-bottom" + (sheetOpen ? " open" : "")}>
        <div className="drawer-head">
          <h3>Фильтры</h3>
          <button className="icon-btn" aria-label="Закрыть" onClick={() => setSheetOpen(false)}><X size={22} /></button>
        </div>
        <div className="drawer-body">
          {renderFilters("m")}
        </div>
        <div className="drawer-foot">
          <button className="btn btn-cta btn-lg btn-block" onClick={() => setSheetOpen(false)}>
            Показать {products.length} {plural(products.length, ["товар", "товара", "товаров"])}
          </button>
        </div>
      </div>
    </>
  );
}

/** URL-запрос из набора фильтров — для сравнения «изменилось ли что-то». */
function toQuery(f: CatalogFilters): string {
  const params = new URLSearchParams();
  if (f.artist) params.set("artist", f.artist);
  if (f.fandom) params.set("fandom", f.fandom);
  if (f.category) params.set("category", f.category);
  if (f.availability !== "all") params.set("availability", f.availability);
  if (f.sort !== "pop") params.set("sort", f.sort);
  return params.toString();
}

function FilterGroup({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="fgroup">
      <h4>{title}</h4>
      {children}
    </div>
  );
}

/** Пункт фильтра: radio (одиночный выбор на измерение). Повторный клик по выбранному снимает выбор. */
function Opt({ name, label, checked, onSelect }: { name: string; label: string; checked: boolean; onSelect: () => void }) {
  return (
    <label className="opt">
      <input
        type="radio"
        name={name}
        checked={checked}
        onChange={() => onSelect()}
        onClick={() => {
          if (checked) onSelect(); // у radio click на выбранном не меняет значение — снимаем вручную
        }}
      />
      <span>{label}</span>
    </label>
  );
}
