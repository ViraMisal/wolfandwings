#!/usr/bin/env bash
# Полный провижинер сервера «Волчица и Крылья» — запускается из клона репозитория
# скриптом-загрузчиком deploy/setup.sh (обычно через curl | sudo bash).
# Напрямую: PROD_USER=vasya PROD_PUBKEY='ssh-ed25519 AAAA...' sudo -E bash provision.sh
#
# СТАТУС: WIP — перед продом требует проверки на чистой копии VPS.
set -euo pipefail

: "${PROD_USER:?задай PROD_USER (--user VASYA)}"
: "${PROD_PUBKEY:?задай PROD_PUBKEY (--key 'ssh-ed25519 AAAA...')}"
APP_DIR="${APP_DIR:-/opt/wolfandwings}"
SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DB_NAME="${DB_NAME:-wolf}"
DB_USER="${DB_USER:-wolf_app}"
DB_PASS="${DB_PASS:-$(openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 28)}"
# пароль попадает и в psql-heredoc, и в DATABASE_URL: никакой спец-пунктуации
case "$DB_PASS" in *[\'/:@?#]* ) echo "DB_PASS не должен содержать ' / : @ ? # — сломает SQL или DSN" >&2; exit 1;; esac
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@wolfandwings.ru}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-$(openssl rand -base64 18 | tr -dc 'A-Za-z0-9' | head -c 20)}"
CFG_DIR="${CFG_DIR:-/etc/wolfandwings}"
SSH_PORT="${SSH_PORT:-22}"

if [ "$(id -u)" -ne 0 ]; then
  echo "запусти от root: sudo -E bash provision.sh" >&2
  exit 1
fi

# PROD_USER попадает в useradd, sudoers-путь и /home — только строгий формат логина
if ! printf '%s' "$PROD_USER" | grep -Eq '^[a-z_][a-z0-9_-]{0,31}$'; then
  echo "PROD_USER: логин в нижнем регистре (a-z, 0-9, _ -), до 32 символов" >&2
  exit 1
fi
# APP_DIR участвует в chown -R: только абсолютный путь без ловушек
case "$APP_DIR" in
  /*) ;;
  *) echo "APP_DIR: ожидаю абсолютный путь" >&2; exit 1 ;;
esac
if [ "$APP_DIR" = "/" ] || [[ "$APP_DIR" == *..* ]] || [[ "$APP_DIR" == *\ * ]]; then
  echo "APP_DIR: подозрительный путь: $APP_DIR" >&2
  exit 1
fi

# предохранитель: провижинер сносит правила фаервола и ставит свои сервисы —
# на «не пустом» сервере без явного FORCE=yes не работаем
if [ ! -d "$CFG_DIR" ] && [ "${FORCE:-}" != "yes" ]; then
  busy=""
  for m in /etc/nginx /etc/apache2 /var/lib/mysql /var/lib/postgresql /var/www; do
    [ -e "$m" ] && busy="$busy $m"
  done
  if command -v ss >/dev/null && ss -tln 2>/dev/null | grep -qE ':(80|443)[[:space:]]'; then
    busy="$busy занятые-порты-80/443"
  fi
  if [ -n "$busy" ]; then
    echo "сервер не выглядит пустым:$busy" >&2
    echo "если осознанно — запусти с FORCE=yes" >&2
    exit 1
  fi
fi

log() { echo "==> $*"; }

install_pkgs() {
  log "apt: postgresql, caddy, fail2ban и утилиты"
  # Caddy — из официального репо (в дефолтных Ubuntu 24.04 его нет)
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' 2>/dev/null | \
    gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg 2>/dev/null || true
  echo "deb [signed-by=/usr/share/keyrings/caddy-stable-archive-keyring.gpg] https://dl.cloudsmith.io/public/caddy/stable/deb/debian any-version main" \
    > /etc/apt/sources.list.d/caddy-stable.list
  apt-get update -y -qq
  apt-get install -y --no-install-recommends \
    postgresql postgresql-contrib caddy fail2ban \
    unattended-upgrades curl ca-certificates \
    openssl git gnupg acl rsync

  log "apt: Node.js 24 (для Next.js standalone)"
  if ! command -v node >/dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_24.x | bash -
    apt-get install -y nodejs
  fi
  echo "node: $(node -v 2>/dev/null || echo 'нет')"
}

create_service_users() {
  log "системные пользователи без shell"
  for u in wolf-api wolf-web; do
    id -u "$u" &>/dev/null || useradd --system --no-create-home --shell /usr/sbin/nologin "$u"
  done
}

create_dirs() {
  log "каталоги приложений"
  install -d -o wolf-api -g wolf-api -m 0750 "$APP_DIR/api" "$APP_DIR/api/data"
  install -d -o wolf-web -g wolf-web -m 0750 "$APP_DIR/web"
  # релизы web: releases/<ts>/ + current — атомарное переключение symlink'ом;
  # бинарник api — один файл, замена mv (wolf-api.new → wolf-api) атомарна
  install -d -o wolf-web -g wolf-web -m 0750 "$APP_DIR/web/releases"
  install -d -o root -g root -m 0750 "$CFG_DIR"
}

write_env_files() {
  log "конфиги в $CFG_DIR"
  cat > "$CFG_DIR/api.env" <<EOF
APP_ENV=prod
LOG_LEVEL=info
ADMIN_HOST=panel.wolfandwings.ru
API_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://${DB_USER}:${DB_PASS}@127.0.0.1:5432/${DB_NAME}?sslmode=disable
WEB_ORIGIN=https://wolfandwings.ru
ADMIN_EMAIL=${ADMIN_EMAIL}
ADMIN_PASSWORD=${ADMIN_PASSWORD}
TG_BOT_TOKEN=
TG_NOTIFY_CHAT=
EOF
  chmod 600 "$CFG_DIR/api.env"

  cat > "$CFG_DIR/web.env" <<EOF
NEXT_PUBLIC_API_URL=/api
API_INTERNAL_URL=http://127.0.0.1:8080
NEXT_PUBLIC_SITE_URL=https://wolfandwings.ru
COMING_SOON=1
EOF
  chmod 600 "$CFG_DIR/web.env"
}

setup_postgres() {
  log "PostgreSQL: роль ${DB_USER} + база ${DB_NAME} (без прав суперпользователя)"
  sudo -u postgres psql -v ON_ERROR_STOP=1 <<SQL
DO \$\$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}') THEN
    CREATE ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASS}';
  ELSE
    ALTER ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASS}';
  END IF;
END \$\$;
SQL
  if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1; then
    sudo -u postgres createdb -O "${DB_USER}" "${DB_NAME}"
  fi
}

harden_ssh() {
  log "SSH: только ключи, без root-логина"
  install -d /etc/ssh/sshd_config.d
  cat > /etc/ssh/sshd_config.d/00-wolf-hardening.conf <<'EOF'
PasswordAuthentication no
PermitRootLogin no
PubkeyAuthentication yes
KbdInteractiveAuthentication no
EOF
  systemctl reload ssh || systemctl reload sshd || true
}

setup_firewall() {
  log "ufw: SSH($SSH_PORT)/80/443, остальное — DROP"
  ufw --force reset 2>/dev/null || true
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow "${SSH_PORT}"/tcp
  ufw allow 22/tcp   # fallback на случай смены порта
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable
  systemctl enable --now fail2ban
}

install_caddy() {
  log "Caddy: конфиг из репозитория"
  install -m 0644 "$SRC_DIR/deploy/caddy/Caddyfile" /etc/caddy/Caddyfile
  systemctl enable caddy
  systemctl reload caddy 2>/dev/null || systemctl restart caddy
  dpkg-reconfigure -f noninteractive unattended-upgrades >/dev/null 2>&1 || true
}

install_units() {
  log "systemd: api, web, watchdog, backup"
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-api.service"      /etc/systemd/system/
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-web.service"      /etc/systemd/system/
  install -m 0755 "$SRC_DIR/deploy/systemd/wolf-watchdog.sh"      /usr/local/lib/wolf-watchdog.sh
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-watchdog.service" /etc/systemd/system/
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-watchdog.timer"   /etc/systemd/system/
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-backup.service"   /etc/systemd/system/
  install -m 0644 "$SRC_DIR/deploy/systemd/wolf-backup.timer"     /etc/systemd/system/
  systemctl daemon-reload
  systemctl enable wolf-api wolf-web
  systemctl enable --now wolf-watchdog.timer wolf-backup.timer
}

setup_admin_user() {
  log "пользователь ${PROD_USER} (sudo, с твоим ключом)"
  id -u "$PROD_USER" &>/dev/null || useradd -m -s /bin/bash "$PROD_USER"
  echo "${PROD_USER} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/"$PROD_USER"
  chmod 0440 /etc/sudoers.d/"$PROD_USER"
  install -d -m 0700 -o "$PROD_USER" -g "$PROD_USER" "/home/$PROD_USER/.ssh"
  echo "$PROD_PUBKEY" > "/home/$PROD_USER/.ssh/authorized_keys"
  chmod 600 "/home/$PROD_USER/.ssh/authorized_keys"
  chown -R "$PROD_USER:$PROD_USER" "/home/$PROD_USER/.ssh"
}

setup_deploy_user() {
  log "пользователь deploy (для GitHub Actions) — узкий sudoers"
  id -u deploy &>/dev/null || useradd -m -s /bin/bash deploy
  usermod -aG wolf-api deploy 2>/dev/null || true
  usermod -aG wolf-web deploy 2>/dev/null || true
  install -d -m 0700 -o deploy -g deploy /home/deploy/.ssh
  touch /home/deploy/.ssh/authorized_keys
  chmod 600 /home/deploy/.ssh/authorized_keys
  chown -R deploy:deploy /home/deploy/.ssh
  # deploy может перезапускать сервисы и chown'ить каталоги под сервисных юзеров
  cat > /etc/sudoers.d/deploy-wolf <<'EOF'
deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart wolf-api, /usr/bin/systemctl restart wolf-web, /usr/bin/systemctl reload caddy
deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl status wolf-api, /usr/bin/systemctl status wolf-api *, /usr/bin/systemctl status wolf-web, /usr/bin/systemctl status wolf-web *
deploy ALL=(root) NOPASSWD: /usr/bin/chown wolf-api.wolf-api /opt/wolfandwings/api/wolf-api, /usr/bin/chown -R wolf-web.wolf-web /opt/wolfandwings/web, /usr/bin/chown -R wolf-api.wolf-api /opt/wolfandwings/api
EOF
  chmod 0440 /etc/sudoers.d/deploy-wolf
  # deploy владеет каталогами приложения (льёт артефакты),
  # сервисные юзеры ходят через группу deploy + ACL
  chown -R deploy:deploy "$APP_DIR"
  usermod -aG deploy wolf-api 2>/dev/null || true
  usermod -aG deploy wolf-web 2>/dev/null || true
  chmod 2775 "$APP_DIR/api" "$APP_DIR/web"
  setfacl -R -m g:deploy:rwx "$APP_DIR/api" "$APP_DIR/web"
  setfacl -d -m g:deploy:rwx "$APP_DIR/api" "$APP_DIR/web"
}

print_summary() {
  cat <<EOF

=========================================================
 ГОТОВО. Дальнейшие шаги:
=========================================================

1) DNS: A-записи wolfandwings.ru, www → IP сервера.
   Алиасы (spiceandwolf.ru, wolfanddice.*, .рф-домены)
   тоже A → IP сервера (Caddy сделает 301 на wolfandwings.ru).
   panel.wolfandwings.ru публичную запись НЕ делать.

2) Учётка админа в Go-API:
   измени ADMIN_PASSWORD в /etc/wolfandwings/api.env на сильный, затем
   sudo systemctl restart wolf-api

3) Доступ к админке — только через SSH-туннель (ничего не торчит наружу):
   - на своей машине добавь в /etc/hosts:  127.0.0.1 panel.wolfandwings.ru
   - подними туннель:  ssh -L 8080:127.0.0.1:8080 ${PROD_USER}@SERVER_IP
   - открой в браузере:  http://panel.wolfandwings.ru:8080/admin/login
   Логин — ADMIN_EMAIL / ADMIN_PASSWORD из api.env.

4) Деплой кода — через GitHub Actions (см. .github/workflows/deploy.yml):
   - в Settings → Secrets добавь: SSH_PRIVATE_KEY, SERVER_HOST, SERVER_USER=deploy,
     и SSH_KNOWN_HOSTS (вывод \`ssh-keyscan -H SERVER_IP\`, фиксирует хост-ключ против MITM)
   - запусти workflow deploy вручную (Actions → deploy → Run workflow).
   - откат web: ln -sfn releases/<прошлый> /opt/wolfandwings/web/current-tmp &&
     mv -T /opt/wolfandwings/web/current-tmp current && sudo systemctl restart wolf-web

Сгенерированные секреты:
  DB_PASS           = ${DB_PASS}
  ADMIN_PASSWORD    = ${ADMIN_PASSWORD}
  (сохранены в ${CFG_DIR}/api.env, chmod 600)
=========================================================
EOF
}

install_pkgs
create_service_users
create_dirs
write_env_files
setup_postgres
harden_ssh
setup_firewall
install_caddy
install_units
setup_admin_user
setup_deploy_user
print_summary
