# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS web-build
WORKDIR /src/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.24.2-alpine AS server-build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY server/ ./server/
COPY sora/ ./sora/
COPY cmd/ ./cmd/
COPY --from=web-build /src/web/dist ./server/dist

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sora2api-server ./server/

FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -h /app appuser

COPY --from=server-build /out/sora2api-server /app/sora2api-server
COPY docker/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh \
	&& chown -R appuser:appuser /app

USER appuser

ENV CONFIG_PATH=/app/config.yaml \
	SERVER_HOST=0.0.0.0 \
	SERVER_PORT=8686 \
	ADMIN_USER=admin \
	ADMIN_PASSWORD=admin123 \
	DATABASE_URL=postgres://postgres:postgres@postgres:5432/sora2api?sslmode=disable \
	NO_PROXY=127.0.0.1,localhost \
	DB_LOG_LEVEL=warn \
	AUTO_MIGRATE=true

EXPOSE 8686

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=5 \
	CMD wget -Y off -qO- http://127.0.0.1:8686/health >/dev/null || exit 1

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/app/sora2api-server"]
