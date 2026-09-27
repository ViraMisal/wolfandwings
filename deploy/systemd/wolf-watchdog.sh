#!/usr/bin/env bash
# Healthcheck wolf-api → Telegram при смене состояния (DOWN/UP).
# Ставится setup.sh в /usr/local/lib/wolf-watchdog.sh, дёргается wolf-watchdog.timer раз в минуту.
# Состояние — в /var/lib/wolf-watchdog/state. Первый запуск (state ещё нет) молчит,
# чтобы после ребута сервера не приходило ложное «UP».
set -euo pipefail

HEALTH_URL="http://127.0.0.1:8080/healthz"
STATE_FILE="/var/lib/wolf-watchdog/state"

if curl -fsS -m 5 "$HEALTH_URL" >/dev/null 2>&1; then
  st="up"
else
  st="down"
fi

prev="unknown"
if [ -f "$STATE_FILE" ]; then
  prev="$(cat "$STATE_FILE")"
fi

if [ "$st" != "$prev" ] && [ "$prev" != "unknown" ]; then
  if [ -n "${TG_BOT_TOKEN:-}" ] && [ -n "${TG_NOTIFY_CHAT:-}" ]; then
    if [ "$st" = "up" ]; then
      text="✅ wolf-api UP"
    else
      text="🩸 wolf-api DOWN"
    fi
    curl -fsS -m 5 -X POST "https://api.telegram.org/bot${TG_BOT_TOKEN}/sendMessage" \
      --data-urlencode "chat_id=${TG_NOTIFY_CHAT}" \
      --data-urlencode "text=${text} @ $(hostname)" >/dev/null \
      || echo "watchdog: не удалось отправить в Telegram" >&2
  fi
fi

echo "$st" > "$STATE_FILE"
