# Runbook — Волчица и Крылья

Что делать, когда «сайт лежит». Команды выполнять на сервере (`ssh prod-user@SERVER_IP`).

> **Статус: WIP** — будет дополняться по мере эксплуатации.

## Первая диагностика (~60 секунд)

```bash
systemctl status wolf-api wolf-web --no-pager   # кто упал
curl -fsS http://127.0.0.1:8080/healthz          # API жив?
curl -fsS -o /dev/null http://127.0.0.1:3000/    # web жив? (200 или 307 при COMING_SOON)
journalctl -u wolf-api -n 100 --no-pager         # хвост логов API
journalctl -u wolf-web -n 100 --no-pager
systemctl list-timers 'wolf-*' --no-pager        # watchdog и бекапы живы?
```

Watchdog сам присылает 🩸 `DOWN` / ✅ `UP` в Telegram, если заполнены `TG_BOT_TOKEN`/`TG_NOTIFY_CHAT` в `/etc/wolfandwings/api.env`.

## Упал после деплоя → откат релиза

```bash
ls -1dt /opt/wolfandwings/web/releases/* | head -3   # какой релиз сейчас
ln -sfn /opt/wolfandwings/web/releases/<прошлый> /opt/wolfandwings/web/current.tmp
mv -T /opt/wolfandwings/web/current.tmp /opt/wolfandwings/web/current
sudo systemctl restart wolf-web
```

API-бинарь хранится в одном экземпляре (замена через `mv`): откат = пересобрать прошлый коммит в Actions и задеплоить повторно.

## Восстановление БД из бекапа

Бекапы: `/var/lib/wolf-backup/wolf-*.dump` (pg_dump custom-формат, 14 дней, ежедневно 04:00).

```bash
sudo systemctl stop wolf-api
set -a; source /etc/wolfandwings/api.env; set +a
pg_restore --clean --if-exists --no-owner -d "$DATABASE_URL" /var/lib/wolf-backup/wolf-<дата>.dump
sudo systemctl start wolf-api
```

**Дрилл (раз в месяц):** восстановить не в боевую, а в пустую базу — бекап, который ни разу не восстанавливали, не бекап:

```bash
sudo -u postgres createdb drill
pg_restore --no-owner -d postgres://wolf_app:<пароль из api.env>@127.0.0.1:5432/drill /var/lib/wolf-backup/wolf-<дата>.dump
sudo -u postgres psql -c "DROP DATABASE drill"
```

## Диски и логи

```bash
journalctl --disk-usage               # журнал растёт — это journald
du -sh /var/lib/wolf-backup           # бекапы
ls /opt/wolfandwings/web/releases     # старые реливы можно снести руками
```

## Конфиги и секреты

- `/etc/wolfandwings/api.env` — API: `ADMIN_PASSWORD`, `TG_*`, `DATABASE_URL` (правишь → `systemctl restart wolf-api`)
- `/etc/wolfandwings/web.env` — web: `COMING_SOON` (правишь → `systemctl restart wolf-web`)

## Метрики

```bash
curl http://127.0.0.1:8080/metrics   # requests, 5xx, goroutines, heap — текст Prometheus
```

Наружу не торчит: API слушает только 127.0.0.1. Prometheus можно снимать через SSH-туннель.
