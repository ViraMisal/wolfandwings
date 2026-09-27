import Link from "next/link";

export const dynamic = "force-dynamic";

export default function NotFound() {
  return (
    <div className="container">
      <div className="empty nf-empty">
        <h1 className="display nf-code">404</h1>
        <p className="muted">Такой страницы нет. Может, она улетела в ночь.</p>
        <Link href="/" className="btn btn-luna btn-lg">На главную</Link>
      </div>
    </div>
  );
}
