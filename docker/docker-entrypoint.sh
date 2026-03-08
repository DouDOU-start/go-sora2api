#!/bin/sh
set -eu

yaml_escape() {
	printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

CONFIG_PATH="${CONFIG_PATH:-/app/config.yaml}"
SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
SERVER_PORT="${SERVER_PORT:-8686}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"
JWT_SECRET="${JWT_SECRET:-}"
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@postgres:5432/sora2api?sslmode=disable}"
DB_LOG_LEVEL="${DB_LOG_LEVEL:-warn}"
AUTO_MIGRATE="${AUTO_MIGRATE:-true}"

mkdir -p "$(dirname "$CONFIG_PATH")"

if [ ! -f "$CONFIG_PATH" ]; then
	echo "[entrypoint] generating config at $CONFIG_PATH"
	umask 077
	{
		printf 'server:\n'
		printf '  host: "%s"\n' "$(yaml_escape "$SERVER_HOST")"
		printf '  port: %s\n' "$SERVER_PORT"
		printf '  admin_user: "%s"\n' "$(yaml_escape "$ADMIN_USER")"
		printf '  admin_password: "%s"\n' "$(yaml_escape "$ADMIN_PASSWORD")"
		if [ -n "$JWT_SECRET" ]; then
			printf '  jwt_secret: "%s"\n' "$(yaml_escape "$JWT_SECRET")"
		fi
		printf '\n'
		printf 'database:\n'
		printf '  url: "%s"\n' "$(yaml_escape "$DATABASE_URL")"
		printf '  log_level: "%s"\n' "$(yaml_escape "$DB_LOG_LEVEL")"
		printf '  auto_migrate: %s\n' "$AUTO_MIGRATE"
	} >"$CONFIG_PATH"
else
	echo "[entrypoint] using existing config at $CONFIG_PATH"
fi

exec "$@"
