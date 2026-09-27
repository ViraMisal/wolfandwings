import type { Metadata } from "next";
import { OrderLookup } from "@/components/OrderLookup";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Статус заказа",
  description: "Проверить статус заказа: номер заказа и email.",
  alternates: { canonical: "/order" },
  robots: { index: false, follow: false },
};

export default function OrderPage() {
  return (
    <div className="container">
      <OrderLookup />
    </div>
  );
}
