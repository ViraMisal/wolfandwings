"use client";
import { useEffect, useRef, useState } from "react";

const STAGES = [
  "Ищем контакты магазина…",
  "Контакты найдены: 1 (одно) видео",
  "Загружаем видеообращение основателей…",
];

/** Фейковый поиск контактов, под конец — видеообращение. Интро с правами
    на музыку пропускается (старт с 2.5 с), звук включается сразу:
    если браузер запретил автозапуск со звуком — по первому клику в любом месте. */
export function RollVideo() {
  const ref = useRef<HTMLVideoElement>(null);
  const [stage, setStage] = useState(0);
  const loading = stage < STAGES.length;

  useEffect(() => {
    if (!loading) return;
    const t = setTimeout(() => setStage((s) => s + 1), stage === 0 ? 1200 : 800);
    return () => clearTimeout(t);
  }, [loading, stage]);

  useEffect(() => {
    if (loading) return;
    const v = ref.current;
    if (!v) return;
    const seek = () => {
      v.currentTime = 2.5;
    };
    if (v.readyState >= 1) seek();
    else v.addEventListener("loadedmetadata", seek, { once: true });

    v.play().catch(() => {
      // автозапуск со звуком не разрешён: запускаем беззвучно, звук — по первому клику
      v.muted = true;
      v.play().catch(() => {});
      const unmute = () => {
        v.muted = false;
        v.play().catch(() => {});
      };
      window.addEventListener("pointerdown", unmute, { once: true });
    });
  }, [loading]);

  if (loading) {
    return (
      <>
        <div className="roll-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </div>
        <p className="muted roll-caption" role="status">
          {STAGES[stage]}
        </p>
      </>
    );
  }

  return (
    <>
      <video ref={ref} className="roll-video" src="/videos/roll.mp4" autoPlay playsInline preload="auto" />
      <p className="muted roll-caption">Рик Эст… а нет, это Искра.</p>
      <p className="subtle roll-contacts">
        Контакты всё-таки есть:{" "}
        <a href="https://t.me/sayarinn" target="_blank" rel="noopener noreferrer">@sayarinn</a>
        {" · "}
        <a href="https://github.com/ViraMisal" target="_blank" rel="noopener noreferrer">github.com/ViraMisal</a>
      </p>
    </>
  );
}
