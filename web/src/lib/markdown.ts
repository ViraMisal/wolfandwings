/** Минимальный и безопасный markdown → HTML: экранирует всё, затем # → h1/h2,
    пустая строка — граница параграфа, https-ссылки — в <a>. */
export function renderMarkdown(md: string): string {
  const esc = md
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  // ссылки собираются ПОСЛЕ экранирования: в url уже нет сырых < и ",
  // поэтому инъекция в атрибут невозможна
  const linkify = (s: string) =>
    s.replace(/https?:\/\/[^\s<]+/g, (m) => {
      const trail = m.match(/[«»"'.;,!?)\]]+$/)?.[0] ?? "";
      const url = trail ? m.slice(0, m.length - trail.length) : m;
      return `<a href="${url}" target="_blank" rel="noopener noreferrer nofollow">${url}</a>${trail}`;
    });
  const lines = esc.split(/\r?\n/);
  const out: string[] = [];
  let para: string[] = [];
  const flush = () => {
    if (para.length) {
      out.push("<p>" + linkify(para.join(" ")) + "</p>");
      para = [];
    }
  };
  for (const raw of lines) {
    const line = raw.trim();
    if (!line) {
      flush();
      continue;
    }
    if (line.startsWith("### ")) {
      flush();
      out.push("<h2>" + line.slice(4) + "</h2>");
    } else if (line.startsWith("## ")) {
      flush();
      out.push("<h2>" + line.slice(3) + "</h2>");
    } else if (line.startsWith("# ")) {
      flush();
      out.push("<h1>" + line.slice(2) + "</h1>");
    } else {
      para.push(line);
    }
  }
  flush();
  return out.join("\n");
}
