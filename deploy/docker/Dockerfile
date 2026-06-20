# Resolve the application version from .git at build time and emit a tiny
# .build-version file. The .git tree is consumed in this throwaway stage and
# never copied into the runtime image.
FROM alpine:3.24 AS version-builder
RUN apk add --no-cache git
WORKDIR /repo
ARG REMNAWAVE_MINISHOP_BRANCH=""
ARG GIT_BRANCH=""
ARG BRANCH_NAME=""
ARG GITHUB_REF_NAME=""
ARG CI_COMMIT_REF_NAME=""
ARG REMNAWAVE_MINISHOP_BUILD_PROVENANCE="custom"
COPY .git ./.git
RUN set -eu; \
    git config --global --add safe.directory /repo; \
    tag=$(git describe --tags --abbrev=0 2>/dev/null || true); \
    sha=$(git rev-parse --short HEAD 2>/dev/null || true); \
    branch="${REMNAWAVE_MINISHOP_BRANCH:-${GIT_BRANCH:-${BRANCH_NAME:-${GITHUB_REF_NAME:-${CI_COMMIT_REF_NAME:-}}}}}"; \
    if [ -z "$branch" ]; then branch=$(git branch --show-current 2>/dev/null || true); fi; \
    case "$branch" in refs/heads/*) branch="${branch#refs/heads/}" ;; origin/*) branch="${branch#origin/}" ;; HEAD) branch="" ;; esac; \
    branch_slug=$(printf '%s' "$branch" | sed -E 's/[^A-Za-z0-9._-]+/-/g; s/^-+//; s/-+$//' | cut -c1-48); \
    branch_suffix=""; \
    if [ -n "$branch_slug" ] && [ "$branch_slug" != "main" ]; then branch_suffix="-$branch_slug"; fi; \
    if [ -n "$tag" ] && [ -n "$sha" ]; then version="${tag}${branch_suffix}+g${sha}"; \
    elif [ -n "$sha" ]; then version="dev${branch_suffix}+g${sha}"; \
    elif [ -n "$tag" ]; then version="${tag}${branch_suffix}"; \
    else version="dev${branch_suffix}+unknown"; fi; \
    provenance=$(printf '%s' "$REMNAWAVE_MINISHOP_BUILD_PROVENANCE" | tr '[:upper:]' '[:lower:]'); \
    case "$provenance" in official|custom|unknown) ;; true|1|yes|upstream|release) provenance="official" ;; false|0|no|fork|modified|local|"") provenance="custom" ;; *) provenance="custom" ;; esac; \
    printf '%s' "$version" > /build-version; \
    printf '%s' "${tag:-unknown}" > /build-tag; \
    printf '%s' "${sha:-unknown}" > /build-commit; \
    printf '%s' "$provenance" > /build-provenance


FROM golang:1.26-bookworm AS go-builder

WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/backend && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate


FROM node:26-slim AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY frontend ./
COPY internal/webassets/templates/subscription_webapp.html ../internal/webassets/templates/subscription_webapp.html
RUN npm run build:webapp


FROM debian:bookworm-slim AS app-base

ENV WEBASSETS_DIR=/app/internal/webassets
WORKDIR /app

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates curl gnupg; \
    . /etc/os-release; \
    install -d /usr/share/postgresql-common/pgdg; \
    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
      | gpg --dearmor -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.gpg; \
    echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.gpg] https://apt.postgresql.org/pub/repos/apt ${VERSION_CODENAME}-pgdg main" \
      > /etc/apt/sources.list.d/pgdg.list; \
    apt-get update; \
    # postgresql-client-17 provides pg_dump and pg_restore for the database
    # backup/restore feature accessible from the admin panel.
    apt-get install -y --no-install-recommends postgresql-client-17; \
    apt-get purge -y --auto-remove gnupg; \
    rm -rf /var/lib/apt/lists/*

RUN useradd -u 10001 -m appuser

COPY locales ./locales
COPY data ./data
COPY internal/webassets ./internal/webassets
COPY --from=frontend-builder /app/internal/webassets/templates/ ./internal/webassets/templates/
COPY --from=go-builder /out/backend /usr/local/bin/remna-user-panel-backend
COPY --from=go-builder /out/worker /usr/local/bin/remna-user-panel-worker
COPY --from=go-builder /out/migrate /usr/local/bin/remna-user-panel-migrate
COPY --from=version-builder /build-version /app/.build-version
COPY --from=version-builder /build-tag /app/.build-tag
COPY --from=version-builder /build-commit /app/.build-commit
COPY --from=version-builder /build-provenance /app/.build-provenance
RUN mkdir -p /app/logs /app/data \
    && chown -R appuser:appuser /app/logs /app/data /app/internal/webassets

USER appuser


FROM app-base AS backend

LABEL org.opencontainers.image.source="https://github.com/remna-user-panel/remna-user-panel" \
      org.opencontainers.image.title="remna-user-panel-backend" \
      org.opencontainers.image.description="Remna User Panel backend (Telegram bot API, web app, webhooks)." \
      org.opencontainers.image.licenses="MIT"

EXPOSE 8080
CMD ["remna-user-panel-backend"]


FROM app-base AS worker

LABEL org.opencontainers.image.source="https://github.com/remna-user-panel/remna-user-panel" \
      org.opencontainers.image.title="remna-user-panel-worker" \
      org.opencontainers.image.description="Remna User Panel background worker (scheduled jobs, async tasks)." \
      org.opencontainers.image.licenses="MIT"

CMD ["remna-user-panel-worker"]
