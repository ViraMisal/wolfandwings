/*
 * Волчица и Крылья — витрина.
 * Copyright (C) 2026 Грёзов Саярин Аквилович (https://github.com/ViraMisal)
 *
 * SPDX-License-Identifier: AGPL-3.0-only
 *
 * Эта программа — свободное ПО: её можно распространять и/или изменять
 * на условиях GNU Affero General Public License v3.0 (исключительно этой
 * версии), опубликованной Free Software Foundation;
 * см. https://www.gnu.org/licenses/agpl-3.0.html и файл LICENSE в корне.
 *
 * Программа распространяется в надежде, что будет полезной, но БЕЗ КАКИХ-ЛИБО
 * ГАРАНТИЙ — без подразумеваемых гарантий товарности и пригодности для
 * конкретных целей. Подробности: https://www.gnu.org/licenses/agpl-3.0.html
 */

import type { Metadata } from "next";
import { Unbounded, Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { LunarBackground } from "@/components/LunarBackground";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import { ToastProvider } from "@/components/Toaster";

const display = Unbounded({
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "600", "700", "800"],
  variable: "--font-display-ui",
  display: "swap",
});
const body = Inter({
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-body-ui",
  display: "swap",
});
const mono = JetBrains_Mono({
  subsets: ["latin", "cyrillic"],
  weight: ["500", "700"],
  variable: "--font-mono-ui",
  display: "swap",
});

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "https://wolfandwings.ru";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "Волчица и Крылья — авторский фан-арт на ковриках",
    template: "%s — Волчица и Крылья",
  },
  description:
    "Магазин авторского фан-арта по Arknights, ZZZ, Honkai: Star Rail и Genshin Impact. Коврики и дескматы от приглашённых художников.",
  applicationName: "Волчица и Крылья",
  icons: { icon: "/brand/favicon.svg" },
  openGraph: {
    type: "website",
    locale: "ru_RU",
    siteName: "Волчица и Крылья",
    title: "Волчица и Крылья — авторский фан-арт на ковриках",
    description:
      "Коврики и дескматы с оригинальным фан-артом по любимым гача-играм. Каждый дизайн — работа приглашённого художника.",
      url: siteUrl,
      images: [],
  },
  robots: { index: true, follow: true },
  alternates: { canonical: "/" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru" className={`${display.variable} ${body.variable} ${mono.variable}`}>
      <body>
        <ToastProvider>
          <LunarBackground />
          <Header />
          <main className="app">{children}</main>
          <Footer />
        </ToastProvider>
      </body>
    </html>
  );
}
