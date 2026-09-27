import type { Metadata } from "next";
import { RollVideo } from "@/components/RollVideo";

export const metadata: Metadata = {
  title: "Наши контакты",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

export default function RollPage() {
  return (
    <div className="container roll-wrap">
      <RollVideo />
    </div>
  );
}
