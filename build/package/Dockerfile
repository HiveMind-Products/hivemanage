# Frontend
FROM node:22-slim AS frontend

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /frontend

COPY web/pnpm-lock.yaml .

RUN pnpm fetch

COPY web .

RUN --mount=type=cache,id=pnpm,target=/pnpm/store CI=true pnpm install --frozen-lockfile
RUN pnpm run build


# Backend
FROM golang:1.26 AS backend

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=frontend /frontend/dist internal/http/dist

ARG LITE_VERSION=dev
ENV CGO_ENABLED=0
RUN go build -ldflags "-X main.Version=${LITE_VERSION}" -o /app/main ./cmd/lite/lite.go

# Application
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=backend /app/main /app/main

COPY --from=backend /app/internal/clickhouse/migrations /app/internal/clickhouse/migrations
COPY --from=backend /app/migrate /app/migrate

EXPOSE 8080

# Run as the built-in non-root user provided by the distroless image.
USER nonroot:nonroot

# No Docker HEALTHCHECK is defined intentionally. This is a distroless/static
# image with no shell and no curl/wget, and the application binary does not
# expose a self-probe subcommand, so any shell- or tool-based HEALTHCHECK would
# be broken. Configure health checks at the orchestrator level (Kubernetes
# probes, ECS, compose, etc.) to hit the HTTP endpoints instead:
#   liveness  -> GET /api/health
#   readiness -> GET /api/ready
# The build context is kept minimal via the repo-root .dockerignore.

CMD ["/app/main"]
