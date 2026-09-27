"use client";
import { useEffect, useRef, useState, type ComponentType, type ReactNode, type Ref } from "react";

/** Появление при скролле в зону видимости. */
export function Reveal({
  children,
  className = "",
  as: Tag = "div",
}: {
  children: ReactNode;
  className?: string;
  as?: "div" | "section" | "ul" | "li";
}) {
  const ref = useRef<HTMLElement>(null);
  const [shown, setShown] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (!("IntersectionObserver" in window)) {
      const id = requestAnimationFrame(() => setShown(true));
      return () => cancelAnimationFrame(id);
    }
    const io = new IntersectionObserver(
      (es) => {
        es.forEach((e) => {
          if (e.isIntersecting) {
            setShown(true);
            io.unobserve(e.target);
          }
        });
      },
      { threshold: 0.08, rootMargin: "0px 0px -6% 0px" }
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  const Comp = Tag as unknown as ComponentType<{ ref?: Ref<HTMLElement>; className?: string; children?: ReactNode }>;
  return (
    <Comp ref={ref} className={"rv" + (shown ? " in" : "") + " " + className}>
      {children}
    </Comp>
  );
}
