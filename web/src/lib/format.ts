const RUB = "\u00A0\u20BD";

/** копейки → "1 290 ₽" */
export function formatPrice(kopecks: number): string {
  return (kopecks / 100).toLocaleString("ru-RU") + RUB;
}

/** склонение по числу: plural(2, ['товар','товара','товаров']) */
export function plural(n: number, forms: [string, string, string]): string {
  n = Math.abs(n) % 100;
  const n1 = n % 10;
  if (n > 10 && n < 20) return forms[2];
  if (n1 > 1 && n1 < 5) return forms[1];
  if (n1 === 1) return forms[0];
  return forms[2];
}

export const STATUS_META: Record<
  string,
  { cls: string; label: string; icon: string }
> = {
  in_stock: { cls: "badge-ok", label: "в наличии", icon: "dot" },
  preorder: { cls: "badge-preorder", label: "предзаказ", icon: "moon" },
  low: { cls: "badge-warn", label: "мало осталось", icon: "flame" },
  archived: { cls: "badge-archived", label: "нет в наличии", icon: "archive" },
};
