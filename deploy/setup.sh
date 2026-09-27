#!/usr/bin/env bash
# Волчица и Крылья — установка VPS одной строкой:
#
#   curl -fsSL https://raw.githubusercontent.com/ViraMisal/wolfandwings/main/deploy/setup.sh | sudo bash -s -- --user vasya --key 'ssh-ed25519 AAAA...'
#
# Это скрипт-загрузчик: он клонирует репозиторий в /opt/wolfandwings/src
# и запускает полный провижинер deploy/provision.sh уже из клона.
# Осторожная альтернатива (сначала посмотреть, потом выполнить):
#
#   curl -fsSL -o setup.sh https://raw.githubusercontent.com/ViraMisal/wolfandwings/main/deploy/setup.sh
#   less setup.sh && sudo bash setup.sh --user vasya --key 'ssh-ed25519 AAAA...'
#
# СТАТУС: WIP — перед продом требует проверки на чистой копии VPS.
set -euo pipefail

REPO="${REPO:-https://github.com/ViraMisal/wolfandwings.git}"
APP_DIR="${APP_DIR:-/opt/wolfandwings}"
BRANCH="${BRANCH:-main}"

usage() {
  cat <<EOF
Установка сервера «Волчица и Крылья» (Ubuntu 24.04 / Debian 12-13, от root).

Использование:
  setup.sh --user VASYA --key 'ssh-ed25519 AAAA...'
  curl -fsSL <url этого скрипта> | sudo bash -s -- --user VASYA --key 'ssh-ed25519 AAAA...'

Параметры (или переменные окружения PROD_USER / PROD_PUBKEY):
  --user  VASYA            логин администратора сервера (получит sudo)
  --key   'ssh-...'        открытый SSH-ключ администратора
  --repo  GIT_URL          репозиторий    (по умолчанию $REPO)
  --branch NAME            ветка          (по умолчанию $BRANCH)
  -h, --help               эта справка
EOF
  exit 0
}

PROD_USER="${PROD_USER:-}"
PROD_PUBKEY="${PROD_PUBKEY:-}"
while [ $# -gt 0 ]; do
  case "$1" in
    --user)  PROD_USER="$2";  shift 2 ;;
    --key)   PROD_PUBKEY="$2"; shift 2 ;;
    --repo)  REPO="$2";       shift 2 ;;
    --branch) BRANCH="$2";    shift 2 ;;
    -h|--help) usage ;;
    *) echo "неизвестный аргумент: $1 (см. --help)" >&2; exit 1 ;;
  esac
done

if [ "$(id -u)" -ne 0 ]; then
  echo "запусти от root: curl ... | sudo bash -s -- --user VASYA --key 'ssh-...'" >&2
  exit 1
fi

# валидация до того, как значения попадут в git/useradd/пути
case "$REPO" in
  https://*|http://*|git@*|file://*|/*) ;;
  *) echo "REPO: ожидаю https://, git@..., file:// или абсолютный путь, получил: $REPO" >&2; exit 1 ;;
esac
case "$APP_DIR" in
  /*) ;;
  *) echo "APP_DIR: ожидаю абсолютный путь" >&2; exit 1 ;;
esac
if ! printf '%s' "$PROD_USER" | grep -Eq '^[a-z_][a-z0-9_-]{0,31}$'; then
  echo "PROD_USER: логин в нижнем регистре (a-z, 0-9, _ -), до 32 символов" >&2
  exit 1
fi
case "$PROD_PUBKEY" in
  ssh-*|ecdsa-*|sk-*) ;;
  *) echo "PROD_PUBKEY: ожидаю открытый SSH-ключ (ssh-ed25519 / ssh-rsa / sk-...)" >&2; exit 1 ;;
esac

if [ -f /etc/os-release ]; then
  . /etc/os-release
  case "${ID:-} ${VERSION_ID:-}" in
    ubuntu*|*debian*) echo "ОС: ${PRETTY_NAME:-неизвестно} — ок" ;;
    *) echo "осторожно: ${PRETTY_NAME:-ОС не определена} — проверялось только на Ubuntu 24.04 / Debian 12-13" >&2 ;;
  esac
fi

if ! command -v git >/dev/null || ! command -v curl >/dev/null; then
  echo "==> ставлю git и curl"
  apt-get update -y -qq
  apt-get install -y --no-install-recommends git curl ca-certificates
fi

SRC_DIR="$APP_DIR/src"
echo "==> клонирую репозиторий в $SRC_DIR"
if [ -d "$SRC_DIR/.git" ]; then
  git -C "$SRC_DIR" fetch --depth 1 origin "$BRANCH"
  git -C "$SRC_DIR" reset --hard "origin/$BRANCH"
else
  git clone --depth 1 -b "$BRANCH" "$REPO" "$SRC_DIR"
fi

echo "==> запускаю провижинер deploy/provision.sh"
export PROD_USER PROD_PUBKEY APP_DIR
exec bash "$SRC_DIR/deploy/provision.sh"
