"use client";
import { useEffect, useRef } from "react";

/** Фиксированный лунный фон: гало + мерцающие звёзды (canvas). Уважает prefers-reduced-motion. */
export function LunarBackground() {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    const reduce = matchMedia("(prefers-reduced-motion: reduce)").matches;
    let w = 0,
      h = 0,
      raf = 0;
    let stars: { x: number; y: number; r: number; a: number; t: number; s: number }[] = [];
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    const init = () => {
      w = canvas.width = innerWidth * dpr;
      h = canvas.height = innerHeight * dpr;
      const count = Math.round((innerWidth * innerHeight) / 11000);
      stars = Array.from({ length: count }, () => ({
        x: Math.random() * w,
        y: Math.random() * h,
        r: Math.random() * 1.2 * dpr + 0.3,
        a: Math.random() * 0.5 + 0.12,
        t: Math.random() * 6.28,
        s: Math.random() * 0.8 + 0.2,
      }));
    };
    init();
    window.addEventListener("resize", init);

    if (reduce) {
      for (const s of stars) {
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r, 0, 7);
        ctx.fillStyle = "rgba(187,212,255," + s.a + ")";
        ctx.fill();
      }
      return () => window.removeEventListener("resize", init);
    }

    const draw = (t: number) => {
      ctx.clearRect(0, 0, w, h);
      for (const s of stars) {
        const tw = 0.6 + 0.4 * Math.sin((t / 900) * s.s + s.t);
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r, 0, 7);
        ctx.fillStyle = "rgba(187,212,255," + s.a * tw + ")";
        ctx.fill();
      }
      raf = requestAnimationFrame(draw);
    };
    raf = requestAnimationFrame(draw);
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("resize", init);
    };
  }, []);

  return (
    <div className="lunar" aria-hidden="true">
      <div className="moon" />
      <div className="moon2" />
      <canvas ref={ref} />
      <div className="grain" />
    </div>
  );
}
